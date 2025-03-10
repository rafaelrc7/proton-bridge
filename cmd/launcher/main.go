// Copyright (c) 2025 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/crash"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v3/internal/versioner"
	"github.com/bradenaw/juniper/xslices"
	"github.com/elastic/go-sysinfo"
	"github.com/elastic/go-sysinfo/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/execabs"
)

const (
	appName      = "Proton Mail Launcher"
	exeName      = "bridge"
	guiName      = "bridge-gui"
	launcherName = "launcher"

	FlagCLI                 = "cli"
	FlagCLIShort            = "c"
	FlagNonInteractive      = "noninteractive"
	FlagNonInteractiveShort = "n"
	FlagLauncher            = "launcher"
	FlagWait                = "wait"
	FlagSessionID           = "session-id"
	HyphenatedFlagLauncher  = "--" + FlagLauncher
	HyphenatedFlagWait      = "--" + FlagWait
	HyphenatedFlagSessionID = "--" + FlagSessionID
)

func main() { //nolint:funlen
	logrus.SetLevel(logrus.DebugLevel)
	l := logrus.WithField("launcher_version", constants.Version)

	reporter := sentry.NewReporter(appName, useragent.New())

	crashHandler := crash.NewHandler(reporter.ReportException)
	defer async.HandlePanic(crashHandler)

	locationsProvider, err := locations.NewDefaultProvider(filepath.Join(constants.VendorName, constants.ConfigName))
	if err != nil {
		l.WithError(err).Fatal("Failed to get locations provider")
	}

	locations := locations.New(locationsProvider, constants.ConfigName)

	logsPath, err := locations.ProvideLogsPath()
	if err != nil {
		l.WithError(err).Fatal("Failed to get logs path")
	}

	sessionID := logging.NewSessionID()
	crashHandler.AddRecoveryAction(logging.DumpStackTrace(logsPath, sessionID, launcherName))

	var closer io.Closer
	if closer, err = logging.Init(
		logsPath,
		sessionID,
		logging.LauncherShortAppName,
		logging.DefaultMaxLogFileSize,
		logging.NoPruning,
		os.Getenv("VERBOSITY"),
	); err != nil {
		l.WithError(err).Fatal("Failed to setup logging")
	}

	defer func() {
		_ = logging.Close(closer)
	}()

	updatesPath, err := locations.ProvideUpdatesPath()
	if err != nil {
		l.WithError(err).Fatal("Failed to get updates path")
	}

	key, err := crypto.NewKeyFromArmored(updater.DefaultPublicKey)
	if err != nil {
		l.WithError(err).Fatal("Failed to create new verification key")
	}

	kr, err := crypto.NewKeyRing(key)
	if err != nil {
		l.WithError(err).Fatal("Failed to create new verification keyring")
	}

	versioner := versioner.New(updatesPath)

	launcher, err := os.Executable()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to determine path to launcher")
	}

	l = l.WithField("launcher_path", launcher)

	args := os.Args[1:]

	exe, err := getPathToUpdatedExecutable(filepath.Base(launcher), versioner, kr)
	if err != nil {
		exeToLaunch := guiName
		if inCLIMode(args) {
			exeToLaunch = exeName
		}

		l = l.WithField("exe_to_launch", exeToLaunch)
		l.WithError(err).Info("No more updates found, looking up bridge executable")

		path, err := versioner.GetExecutableInDirectory(exeToLaunch, filepath.Dir(launcher))
		if err != nil {
			l.WithError(err).Fatal("No executable in launcher directory")
		}

		exe = path
	}

	l = l.WithField("exe_path", exe)

	args, wait, mainExes := findAndStripWait(args)
	if wait {
		for _, mainExe := range mainExes {
			waitForProcessToFinish(mainExe)
		}
	}

	cmd := execabs.Command(exe, appendLauncherPath(launcher, appendOrModifySessionID(args, string(sessionID)))...) //nolint:gosec

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	// On windows, if you use Run(), a terminal stays open; we don't want that.
	if //goland:noinspection GoBoolExpressions
	runtime.GOOS == "windows" {
		err = cmd.Start()
	} else {
		err = cmd.Run()
	}

	if err != nil {
		l.WithError(err).Fatal("Failed to launch")
	}
}

// appendLauncherPath add launcher path if missing.
func appendLauncherPath(path string, args []string) []string {
	if !slices.Contains(args, HyphenatedFlagLauncher) {
		res := append([]string{}, args...)
		res = append(res, HyphenatedFlagLauncher, path)
		return res
	}
	return args
}

// inCLIMode detect if CLI mode is asked.
func inCLIMode(args []string) bool {
	return hasFlag(args, FlagCLI) || hasFlag(args, FlagCLIShort) || hasFlag(args, FlagNonInteractive) || hasFlag(args, FlagNonInteractiveShort)
}

// hasFlag checks if a flag is present in a list.
func hasFlag(args []string, flag string) bool {
	return flagIndex(args, flag) >= 0
}

// flagIndex returns the position of the first occurrence of a flag int args, or -1 if the flag is not present.
func flagIndex(args []string, flag string) int {
	return slices.IndexFunc(args, func(arg string) bool { return (arg == "-"+flag) || (arg == "--"+flag) })
}

// findAndStrip check if a value is present in s list and remove all occurrences of the value from this list.
func findAndStrip[T comparable](slice []T, v T) (strippedList []T, found bool) {
	strippedList = xslices.Filter(slice, func(value T) bool {
		return value != v
	})
	return strippedList, len(strippedList) != len(slice)
}

// findAndStripWait Check for waiter flag get its value and clean them both.
func findAndStripWait(args []string) ([]string, bool, []string) {
	res := append([]string{}, args...)

	hasFlag := false
	values := make([]string, 0)
	for k, v := range res {
		if v != HyphenatedFlagWait {
			continue
		}
		if k+1 >= len(res) {
			continue
		}
		hasFlag = true
		values = append(values, res[k+1])
	}

	if hasFlag {
		res, _ = findAndStrip(res, HyphenatedFlagWait)
		for _, v := range values {
			res, _ = findAndStrip(res, v)
		}
	}
	return res, hasFlag, values
}

// return args with the sessionID flag and value added or modified. The original slice is not modified.
func appendOrModifySessionID(args []string, sessionID string) []string {
	index := flagIndex(args, FlagSessionID)
	if index < 0 {
		return append(args, HyphenatedFlagSessionID, sessionID)
	}

	if index == len(args)-1 {
		return append(args, sessionID)
	}

	res := slices.Clone(args)
	res[index+1] = sessionID

	return res
}

func getPathToUpdatedExecutable(
	name string,
	ver *versioner.Versioner,
	kr *crypto.KeyRing,
) (string, error) {
	versions, err := ver.ListVersions()
	if err != nil {
		return "", errors.Wrap(err, "failed to list available versions")
	}

	currentVersion, err := semver.StrictNewVersion(constants.Version)
	if err != nil {
		logrus.WithField("version", constants.Version).WithError(err).Error("Failed to parse current version")
	}

	for _, version := range versions {
		vlog := logrus.WithFields(logrus.Fields{
			"version":       constants.Version,
			"check_version": version,
			"name":          name,
		})

		if err := version.VerifyFiles(kr); err != nil {
			vlog.WithError(err).Error("Files failed verification and will be removed")

			if err := version.Remove(); err != nil {
				vlog.WithError(err).Error("Failed to remove files")
			}

			continue
		}

		// Skip versions that are less or equal to launcher version.
		if currentVersion != nil && !version.SemVer().GreaterThan(currentVersion) {
			continue
		}

		exe, err := version.GetExecutable(name)
		if err != nil {
			vlog.WithError(err).Error("Failed to get executable")
			continue
		}

		return exe, nil
	}

	return "", errors.New("no available newer versions")
}

// waitForProcessToFinish waits until the process with the given path is finished.
func waitForProcessToFinish(exePath string) {
	for {
		processes, err := sysinfo.Processes()
		if err != nil {
			logrus.WithError(err).Error("Could not determine running processes")
			return
		}

		exeInfo, err := os.Stat(exePath)
		if err != nil {
			logrus.WithError(err).WithField("file", exeInfo).Error("Could not retrieve file info")
			return
		}

		if xslices.Any(processes, func(process types.Process) bool {
			info, err := process.Info()
			if err != nil {
				logrus.WithError(err).Trace("Could not retrieve process info")
				return false
			}

			return sameFile(exeInfo, info.Exe)
		}) {
			logrus.Infof("Waiting for %v to finish.", exeInfo.Name())
			time.Sleep(1 * time.Second)
			continue
		}

		return
	}
}

func sameFile(info os.FileInfo, path string) bool {
	pathInfo, err := os.Stat(path)
	if err != nil {
		logrus.WithError(err).WithField("file", path).Error("Could not retrieve file info")
		return false
	}

	return os.SameFile(pathInfo, info)
}
