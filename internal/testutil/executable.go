package testutil

import (
	"os"
	"syscall"
)

// WriteExecutable is os.WriteFile for scripts a test is about to run.
//
// A parallel test that forks between this file being opened for writing and closed hands the
// child a copy of the write descriptor until the child execs; running the script meanwhile fails
// with "text file busy" (ETXTBSY, golang/go#22315). Holding syscall.ForkLock for reading keeps
// os/exec from forking while the descriptor is open.
func WriteExecutable(name string, data []byte, perm os.FileMode) error {
	syscall.ForkLock.RLock()
	defer syscall.ForkLock.RUnlock()
	return os.WriteFile(name, data, perm)
}
