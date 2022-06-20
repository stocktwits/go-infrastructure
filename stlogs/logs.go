package stlogs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
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
var localLoggers map[string]Logger

func init() {
	localLoggers = make(map[string]Logger)
}

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
	testLevel(level string, msg string) ([]byte, error)
}

//An audit logger, this is a singleton and implements the Logger interface
type AuditLogger struct {
	logger   *logrus.Logger
	app      string
	hostname string
}

//Data that will be fw using the context
type InfoCtx struct {
	auditData map[string]interface{}
	auditTags []string
}

//A new log entry, this is a log entry to be printed, include commond fields
type AuditEntry struct {
	*logrus.Entry
	auditLogger *AuditLogger
	*InfoCtx
}

//Json formater
type STJSONFormater struct {
	logrus.JSONFormatter
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

	return []byte(sdata), nil
}

//This creates a new AuditEntry object with the same values as the original one
//The modifications done to this entry will not be preserved in the other logs
func (ae *AuditEntry) NewEntry() Logger {
	newEntry := ae.auditLogger.newAuditEntry()

	for k, v := range ae.auditData {
		newEntry.AddData(k, v)
	}

	for _, t := range ae.auditTags {
		newEntry.AddTag(t)
	}

	return newEntry
}

//Links a logger with a context, this is usefull to keep using the same auditEntry
//in different parts of the application where the context is passed
func newWithContext(ctx context.Context, ae *AuditEntry) (Logger, context.Context) {
	var newCtx context.Context

	if infCtx, ok := ctx.Value(InfoCtxKey).(*InfoCtx); ok {
		nae := ae.auditLogger.newAuditEntry()
		nae.InfoCtx = infCtx
		lock.Lock()
		if len(nae.auditData) > 0 {
			nae.Entry = nae.WithField("data", nae.auditData)
		}
		if len(nae.auditTags) > 0 {
			nae.Entry = nae.WithField("tags", nae.auditTags)
		}
		lock.Unlock()

		return nae, ctx
	} else {
		ae.auditData = map[string]interface{}{}
		ae.auditTags = make([]string, 0)

		ae.AddData("txId", getID())
		lock.Lock()
		newCtx = context.WithValue(ctx, InfoCtxKey, ae.InfoCtx)
		lock.Unlock()

		return ae, newCtx
	}

}

//Links a logger with a context from an AuditEntry
func (ae *AuditEntry) NewWithContext(ctx context.Context) (Logger, context.Context) {
	return newWithContext(ctx, ae)
}

//Adds a new entry to the data map
//This value will be printed in all the logs that uses the same entry object, or that uses the same context
func (ae *AuditEntry) AddData(key string, value interface{}) Logger {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := ae.Data["data"]; !ok {
		ae.Entry = ae.WithField("data", ae.auditData)
	}

	ae.auditData[key] = value

	return ae

}

//Adds a new tag to the tags array
//This value will be printed in all the logs that uses the same entry object, or that uses the same context
func (ae *AuditEntry) AddTag(tag string) Logger {
	lock.Lock()
	defer lock.Unlock()

	ae.auditTags = append(ae.auditTags, tag)

	if _, ok := ae.Data["tags"]; !ok {
		ae.Entry = ae.WithField("tags", ae.auditTags)
	} else {
		ae.Data["tags"] = ae.auditTags
	}

	return ae
}

func (ae *AuditEntry) AddTags(tags ...string) Logger {
	for _, tag := range tags {
		ae.AddTag(tag)
	}

	return ae
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
	log := ae.NewEntry()
	for _, tag := range tags {
		log = log.AddTag(tag)
	}
	return log
}

//Creates an error entry in the data map
func (ae *AuditEntry) WithError(e error) Logger {
	if e == nil {
		e = fmt.Errorf("nil error was logged")
	}
	return ae.WithData("error", e.Error())
}

func (al *AuditLogger) newAuditEntry() *AuditEntry {

	entry := al.logger.WithField("src", al.app)

	entry = entry.WithField("host", al.hostname)

	entry = entry.WithField("sv", SchemaVersion)

	return &AuditEntry{
		Entry:       entry,
		auditLogger: al,
		InfoCtx: &InfoCtx{
			auditData: map[string]interface{}{},
			auditTags: []string{},
		},
	}

}

func getFormat() *STJSONFormater {
	return &STJSONFormater{logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "ts",
		},
		PrettyPrint: prettyPrint,
	}}
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

	format := STJSONFormater{logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "ts",
		},
	}}

	singleLogger.logger.SetFormatter(&format)

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

	al.logger.SetFormatter(getFormat())
	al.logger.SetOutput(al.logger.Out)

	return al
}

//Creates a new Local Logger, it copies the information from the global one.
//If the global one is not created the information will be set as UNDEFINED
func NewLocal(module string) Logger {
	if logger, ok := localLoggers[module]; ok {
		return logger
	}

	al := newAuditLogger(module)
	return al.newAuditEntry()
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
