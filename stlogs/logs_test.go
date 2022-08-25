package stlogs

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

type Log struct {
	Id   string                 `json:"id"`
	Lv   int                    `json:"lv"`
	Src  string                 `json:"src"`
	Host string                 `json:"host"`
	Msg  string                 `json:"msg"`
	Ts   time.Time              `json:"ts"`
	Tags []string               `json:"tags"`
	Data map[string]interface{} `json:"data"`
}

func TestSingleton(t *testing.T) {
	t.Parallel()

	log1 := NewGlobal("debug", "test")

	log2 := NewGlobal("info", "other info")

	if log1 == log2 {
		t.Error("returning the same entry")
	}

	log3 := NewLocal("module")

	if log1 == log3 {
		t.Error("local logger was not created as a new instance")
	}

}

func TestValues(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		level    string
		intLevel Level
		msg      string
		tags     []string
		data     map[string]interface{}
	}{
		{
			level:    "debug",
			intLevel: 10,
			msg:      "debug test",
			tags:     []string{"tag1", "tag2"},
			data:     map[string]interface{}{"testdata": "value"},
		},
		{
			level:    "info",
			intLevel: 20,
			msg:      "info test",
			tags:     []string{"tag1", "tag2"},
			data:     map[string]interface{}{"testdata": "value"},
		},
		{
			level:    "warning",
			intLevel: 30,
			msg:      "warning test",
			tags:     []string{"tag1", "tag2"},
			data:     map[string]interface{}{"testdata": "value"},
		},
		{
			level:    "error",
			intLevel: 40,
			msg:      "error test",
			tags:     []string{"tag1", "tag2"},
			data:     map[string]interface{}{"testdata": "value"},
		},
		{
			level:    "fatal",
			intLevel: 50,
			msg:      "fatal test",
			tags:     []string{"tag1", "tag2"},
			data:     map[string]interface{}{"testdata": "value"},
		},
	}

	logs := NewGlobal("debug", "test")

	for _, test := range testTable {
		tnow := time.Now()

		logSt := Log{}

		for k, v := range test.data {
			logs.AddData(k, v)
		}

		for _, tag := range test.tags {
			logs.AddTag(tag)
		}

		data, err := logs.testLevel(test.level, test.msg)

		if err != nil {
			t.Errorf("error while running %s: %v", test.msg, err)
			continue
		}

		if len(data) == 0 {
			t.Errorf("error while running %s, no data", test.msg)
			continue
		}

		err = json.Unmarshal(data, &logSt)

		if err != nil {
			t.Errorf("error while running %s: %v", test.msg, err)
			continue
		}

		if len(logSt.Id) != 26 {
			t.Error("Wrong id length")
		}

		if logSt.Lv != int(test.intLevel) {
			t.Errorf("Error value not set properly, have: %d, want: %d", logSt.Lv, test.intLevel)
		}

		if logSt.Msg != test.msg {
			t.Errorf("Wrong message value, have: %s, want: %s", logSt.Msg, test.msg)
		}

		diff := tnow.Sub(logSt.Ts).Milliseconds()
		if diff > 1000 {
			t.Errorf("wrong time stampt, have: %v, want: %v", logSt.Ts, tnow)
		}

		if logSt.Src != "test" {
			t.Errorf("wrong soruce, want: test0, have: %s", logSt.Src)
		}

		hn, err := os.Hostname()

		if err != nil {
			t.Error("error getting the host name from OS")
		}

		if logSt.Host != hn {
			t.Errorf("wrong host name, want: %s, have: %s", hn, logSt.Host)
		}

		if len(logSt.Data) != len(test.data) {
			t.Error("missing data information")
		}

		for k, v := range logSt.Data {
			if rv, ok := test.data[k]; !ok || rv != v {
				t.Error("wrong data values")
			}
		}

	}
}

func TestLocalLogger(t *testing.T) {
	t.Parallel()

	NewGlobal("debug", "test")

	logs := NewLocal("local")

	logSt := Log{}

	data, err := logs.testLevel("debug", "test local")

	if err != nil {
		t.Errorf("error will running log: %v", err)
	}

	_ = json.Unmarshal(data, &logSt)

	if logSt.Src != "test/local" {
		t.Errorf("wrong source name, want test/local, have: %v, data: %s", logSt, string(data))
	}
}

func TestLocalLoggerWithLevel(t *testing.T) {

	NewGlobal("debug", "test")

	logs := NewLocalWithLevel("with_level", "info")

	logSt := Log{}

	data, err := logs.testLevel("debug", "test with level")

	if err != nil {
		t.Errorf("error will running log: %v", err)
	}

	_ = json.Unmarshal(data, &logSt)

	if logSt.Src != "test/with_level" {
		t.Errorf("wrong source name, want test/with_level, have: %s", logSt.Src)
	}
}

func OtherPlace(ctx context.Context) {
	log := NewLocal("module")
	log, _ = log.NewWithContext(ctx)

	log.AddData("test_ctx2", "data2")
}

func TestWithContext(t *testing.T) {
	t.Parallel()

	NewGlobal("debug", "test")

	ctx := context.Background()

	log := NewGlobal("debug", "test")

	log, ctx = log.NewWithContext(ctx)

	log.AddData("test_ctx1", "data1")

	OtherPlace(ctx)

	logSt := Log{}

	data, err := log.testLevel("debug", "test with context")

	if err != nil {
		t.Errorf("error will running log: %v", err)
	}

	_ = json.Unmarshal(data, &logSt)

	value, ok := logSt.Data["test_ctx1"]
	if !ok {
		t.Fatalf("value passed on context not found")
	}

	if value != "data1" {
		t.Errorf("wrong value recoverd from context, want 'data1', have: %s", value)
	}

	value2, ok := logSt.Data["test_ctx2"]
	if !ok {
		t.Fatalf("value passed on context now found")
	}

	if value2 != "data2" {
		t.Errorf("wrong value recoverd from context, want 'data2', have: %s", value2)
	}

}

func TestConcurrency(t *testing.T) {
	logger := NewLocal("test-concurrency")

	for i := 0; i < 0; i++ {
		ctx := context.Background()
		log, ctx := logger.NewWithContext(ctx)
		log.AddTag("tag")

		log.Info("test 1")

		testPrint(ctx)

		log.Info("test 3")
	}
}

func testPrint(ctx context.Context) {
	logger := NewLocal("test-print")
	log, _ := logger.NewWithContext(ctx)
	log.AddTag("tag")

	log.Info("test 2")
}
