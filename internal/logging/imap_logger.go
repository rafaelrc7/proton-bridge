// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package logging

import "github.com/sirupsen/logrus"

// IMAPLogger implements the writer interface for Gluon IMAP logs
type IMAPLogger struct {
	l *logrus.Entry
}

func NewIMAPLogger() *IMAPLogger {
	return &IMAPLogger{l: logrus.WithField("pkg", "IMAP")}
}

func (l *IMAPLogger) Write(p []byte) (n int, err error) {
	return l.l.WriterLevel(logrus.TraceLevel).Write(p)
}