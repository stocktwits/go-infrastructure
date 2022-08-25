package stlogs

func (ae *AuditEntry) Tracef(format string, args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Tracef(format, args...)
}

func (ae *AuditEntry) Debugf(format string, args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Debugf(format, args...)
}

func (ae *AuditEntry) Infof(format string, args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Infof(format, args...)
}

func (ae *AuditEntry) Warnf(format string, args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Warnf(format, args...)
}

func (ae *AuditEntry) Errorf(format string, args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Errorf(format, args...)
}

func (ae *AuditEntry) Fatalf(format string, args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Fatalf(format, args...)
}

func (ae *AuditEntry) Trace(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Trace(args...)
}

func (ae *AuditEntry) Debug(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Debug(args...)
}

func (ae *AuditEntry) Info(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Info(args...)
}

func (ae *AuditEntry) Warn(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Warn(args...)
}

func (ae *AuditEntry) Error(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Error(args...)
}

func (ae *AuditEntry) Fatal(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Fatal(args...)
}

func (ae *AuditEntry) Traceln(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Traceln(args...)
}

func (ae *AuditEntry) Debugln(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Debugln(args...)
}

func (ae *AuditEntry) Infoln(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Infoln(args...)
}

func (ae *AuditEntry) Warnln(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Warnln(args...)
}

func (ae *AuditEntry) Errorln(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Errorln(args...)
}

func (ae *AuditEntry) Fatalln(args ...interface{}) {
	ae.Lock()
	defer ae.Unlock()

	ae.getEntry().Fatalln(args...)
}
