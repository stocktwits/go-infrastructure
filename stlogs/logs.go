package stlogs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid"
	"github.com/sirupsen/logrus"
)

const SchemaVersion = 1

type LogCtxKey uint8

const (
	InfoCtxKey = LogCtxKey(iota)
)

//Defines a level
type Level int

//Level values
const (
	DEBUG = Level(10 * (iota + 1))
	INFO
	WARN
	ERROR
	FATAL
)

//Var global mutex
var lock sync.Mutex

//The AuditLogger is created as a sigleton
var singleLogger *AuditLogger

//Defines if logs will be printed pretty
var prettyPrint bool

//Local loggers
var localLoggers map[string]*AuditLogger = make(map[string]*AuditLogger)

//Converts the logrus levels into local levels
func getLevel(level string) Level {
	switch level {
	case "fatal":
		return FATAL
	case "error":
		return ERROR
	case "warning":
		return WARN
	case "info":
		return INFO
	case "debug":
		return DEBUG
	case "trace":
		return DEBUG
	default:
		return INFO
	}
}

//Set Pretty flag
func SetPretty(f bool) {
	prettyPrint = f
}

//Generates a new log ID
func getID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)

	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

//This interface was added to limit some unneeded log functions

type LogPrinter interface {
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Traceln(args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
}

//Interface defines a logger, this must be used in all implementations
type Logger interface {
	LogPrinter
	AddTag(string) Logger
	AddTags(...string) Logger
	AddData(key string, value interface{}) Logger
	WithTag(string) Logger
	WithTags(...string) Logger
	WithData(key string, value interface{}) Logger
	WithError(err error) Logger
	NewEntry() Logger
	NewWithContext(ctx context.Context) (Logger, context.Context)
	AddSensitive(...string)
	testLevel(level string, msg string) ([]byte, error)
}

//An audit logger, this is a singleton and implements the Logger interface
type AuditLogger struct {
	logger    *logrus.Logger
	app       string
	hostname  string
	sensitive []string
}

//Data that will be fw using the context
type InfoCtx struct {
	auditData map[string]interface{}
	auditTags []string
}

//A new log entry, this is a log entry to be printed, include commond fields
type AuditEntry struct {
	auditLogger *AuditLogger
	info        *InfoCtx
	sync.RWMutex
}

//Json formater
type STJSONFormater struct {
	logrus.JSONFormatter
	re *regexp.Regexp
}

func newSTJSONFormater(sensitive []string) *STJSONFormater {
	var re *regexp.Regexp
	if len(sensitive) > 0 {
		qSensitive := []string{}
		for _, s := range sensitive {
			qSensitive = append(qSensitive, regexp.QuoteMeta(s))
		}

		keys := strings.Join(qSensitive, "|")
		reString := fmt.Sprintf(`((\"|\')(%s)(\"|\'):\s*)((\"|\')([^\"\']+)(\"|\')|[\d\.]+)`, keys)
		re = regexp.MustCompile(reString)
	}

	return &STJSONFormater{
		JSONFormatter: logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime: "ts",
			},
			PrettyPrint: prettyPrint,
		},
		re: re,
	}
}

//Re-implements Formater to change log level format
func (f *STJSONFormater) Format(entry *logrus.Entry) ([]byte, error) {
	slv := entry.Level.String()

	bdata, err := f.JSONFormatter.Format(entry)

	if err != nil {
		return nil, err
	}

	sdata := string(bdata)

	if prettyPrint {
		sdata = strings.Replace(sdata, "  \"level\": \""+slv+"\",\n", "", 1)
	} else {
		sdata = strings.Replace(sdata, "\"level\":\""+slv+"\",", "", 1)
	}

	if f.re != nil {
		sdata = f.re.ReplaceAllString(sdata, `${1}"****"`)
	}

	return []byte(sdata), nil
}

func (ae *AuditEntry) copyInfo() *InfoCtx {
	ae.Lock()
	defer ae.Unlock()

	newData := map[string]interface{}{}
	newTags := []string{}

	for k, v := range ae.info.auditData {
		newData[k] = v
	}

	newTags = append(newTags, ae.info.auditTags...)

	return &InfoCtx{
		auditData: newData,
		auditTags: newTags,
	}
}

func (ae *AuditEntry) getEntry() *logrus.Entry {
	al := ae.auditLogger

	entry := al.logger.WithField("src", al.app)

	entry = entry.WithField("host", al.hostname)

	entry = entry.WithField("sv", SchemaVersion)

	if len(ae.info.auditData) > 0 {
		entry = entry.WithField("data", ae.info.auditData)
	}

	if len(ae.info.auditTags) > 0 {
		entry = entry.WithField("tags", ae.info.auditTags)
	}

	return entry
}

func (al *AuditLogger) newAuditEntry() *AuditEntry {
	return &AuditEntry{
		auditLogger: al,
		info: &InfoCtx{
			auditData: map[string]interface{}{},
			auditTags: []string{},
		},
	}
}

//This creates a new AuditEntry object with the same values as the original one
//The modifications done to this entry will not be preserved in the other logs
func (ae *AuditEntry) NewEntry() Logger {
	newEntry := ae.auditLogger.newAuditEntry()

	info := ae.copyInfo()

	for k, v := range info.auditData {
		newEntry.AddData(k, v)
	}

	for _, t := range info.auditTags {
		newEntry.AddTag(t)
	}

	return newEntry
}

//Links a logger with a context from an AuditEntry
func (ae *AuditEntry) NewWithContext(ctx context.Context) (Logger, context.Context) {
	var newCtx context.Context

	nae := ae.auditLogger.newAuditEntry()

	if infCtx, ok := ctx.Value(InfoCtxKey).(*InfoCtx); ok {
		nae.info = infCtx
		return nae, ctx
	} else {
		nae.AddData("txId", getID())

		nae.Lock()
		newCtx = context.WithValue(ctx, InfoCtxKey, nae.info)
		nae.Unlock()

		return nae, newCtx
	}
}

//Adds a key to be recognized as sensitive data. This will use for maps keys and structures field names
func (ae *AuditEntry) AddSensitive(s ...string) {
	ae.auditLogger.sensitive = append(ae.auditLogger.sensitive, s...)
	ae.auditLogger.logger.SetFormatter(newSTJSONFormater(ae.auditLogger.sensitive))
}

//Adds a new entry to the data map
//This value will be printed in all the logs that uses the same entry object, or that uses the same context
func (ae *AuditEntry) AddData(key string, value interface{}) Logger {
	ae.Lock()
	defer ae.Unlock()

	ae.info.auditData[key] = value

	return ae

}

//Adds a new tag to the tags array
//This value will be printed in all the logs that uses the same entry object, or that uses the same context
func (ae *AuditEntry) AddTag(tag string) Logger {
	ae.Lock()
	defer ae.Unlock()

	ae.info.addTags(tag)

	return ae
}

func (ae *AuditEntry) AddTags(tags ...string) Logger {
	ae.Lock()
	defer ae.Unlock()

	ae.info.addTags(tags...)

	return ae
}

func (i *InfoCtx) addTags(tags ...string) {
	new := append(i.auditTags, tags...)
	set := make(map[string]interface{})

	for _, t := range new {
		set[t] = nil
	}

	reduced := []string{}
	for k := range set {
		reduced = append(reduced, k)
	}

	i.auditTags = reduced
}

//Creates a new entry and adds the give value to the data map
//This value will not be printed in other logs
func (ae *AuditEntry) WithData(key string, value interface{}) Logger {
	return ae.NewEntry().AddData(key, value)
}

//Creates a new entry and adds the give value to the tags array
//This value will not be printed in other logs
func (ae *AuditEntry) WithTag(tag string) Logger {
	return ae.NewEntry().AddTag(tag)
}

//Allows to add multiple tags
//This value will not be printed in other logs
func (ae *AuditEntry) WithTags(tags ...string) Logger {
	return ae.NewEntry().AddTags(tags...)
}

//Creates an error entry in the data map
func (ae *AuditEntry) WithError(e error) Logger {
	if e == nil {
		e = fmt.Errorf("nil error was logged")
	}
	return ae.WithData("error", e.Error())
}

//Creates a new global logger, this is singleton
//if you call this function twice, it will return an AuditLogger with
//the information of the the first call.
func NewGlobal(level string, app string) Logger {
	lock.Lock()
	defer lock.Unlock()

	if singleLogger != nil {
		return singleLogger.newAuditEntry()
	}

	//Set hostname
	hn, err := os.Hostname()
	if err != nil {
		hn = "UNDEFINED"
	}

	singleLogger = &AuditLogger{
		logger:   logrus.New(),
		app:      app,
		hostname: hn,
	}

	logrusLevel, err := logrus.ParseLevel(level)

	if err != nil {
		logrusLevel = logrus.InfoLevel
	}

	singleLogger.logger.SetLevel(logrusLevel)

	singleLogger.logger.AddHook(&PrintHook{})

	singleLogger.logger.SetFormatter(newSTJSONFormater(nil))

	return singleLogger.newAuditEntry()
}

func newAuditLogger(module string) *AuditLogger {
	al := &AuditLogger{
		logger: logrus.New(),
	}

	if singleLogger != nil {
		al.app = singleLogger.app + "/" + module
		al.hostname = singleLogger.hostname
		al.logger.SetLevel(singleLogger.logger.Level)
	} else {
		//Set hostname
		hn, err := os.Hostname()
		if err != nil {
			hn = "UNDEFINED"
		}
		al.hostname = hn
		al.app = module
	}

	al.logger.AddHook(&PrintHook{})

	al.logger.SetFormatter(newSTJSONFormater(nil))
	al.logger.SetOutput(al.logger.Out)

	return al
}

//Creates a new Local Logger, it copies the information from the global one.
//If the global one is not created the information will be set as UNDEFINED
func NewLocal(module string) Logger {
	lock.Lock()
	defer lock.Unlock()

	if logger, ok := localLoggers[module]; ok {
		return logger.newAuditEntry()
	}

	localLoggers[module] = newAuditLogger(module)

	return localLoggers[module].newAuditEntry()
}

//This allows you to create a local copy of the Global Logger
//If you define a log level lower than the global one, the new log level will be applied
func NewLocalWithLevel(module string, level string) Logger {
	al := newAuditLogger(module)

	logrusLevel, err := logrus.ParseLevel(level)

	if err != nil && singleLogger != nil {
		if singleLogger != nil {
			logrusLevel = singleLogger.logger.Level
		} else {
			logrusLevel = logrus.InfoLevel
		}
	}

	if singleLogger != nil && singleLogger.logger.Level > logrusLevel {
		logrusLevel = singleLogger.logger.Level
	}

	al.logger.SetLevel(logrusLevel)

	return al.newAuditEntry()

}

type PrintHook struct{}

func (ph *PrintHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (ph *PrintHook) Fire(entry *logrus.Entry) error {
	slv := entry.Level.String()

	entry.Data["lv"] = int(getLevel(slv))

	entry.Data["id"] = getID()

	return nil
}

/* Testing */

type loggerMap map[string]func(...interface{})

func getLoggerMap(l Logger) loggerMap {
	fs := make(loggerMap)

	fs["debug"] = l.Debug
	fs["warning"] = l.Warn
	fs["error"] = l.Error
	fs["info"] = l.Info
	fs["fatal"] = l.Fatal

	return fs
}

//This function was created to do testing
//It deactive the exit function and return the log result in a bytes stream
func (ae *AuditEntry) testLevel(level string, msg string) ([]byte, error) {
	lock.Lock()
	defer lock.Unlock()

	tmp := ae.auditLogger.logger.ExitFunc

	ae.auditLogger.logger.ExitFunc = func(int) {}

	buf := bytes.NewBuffer([]byte{})

	ae.auditLogger.logger.SetOutput(buf)

	fs := getLoggerMap(ae)

	fs[level](msg)

	data, err := ioutil.ReadAll(buf)

	ae.auditLogger.logger.SetOutput(os.Stderr)
	ae.auditLogger.logger.ExitFunc = tmp

	return data, err
}
