package kb

import (
	"os"
	"syscall"
)

// LockWrites serializes file mutations and cache transactions across processes.
// The returned release function must run after reindexing and any git commit.
func LockWrites(root string) (func(), error) {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, refuse(err)
	}
	defer dir.Close()
	stat, err := dir.Lstat(".mdkb")
	if err != nil {
		return nil, refuse(err)
	}
	if !stat.IsDir() || stat.Mode()&os.ModeSymlink != 0 {
		return nil, refuse(os.ErrPermission)
	}
	f, err := dir.OpenFile(".mdkb/write.lock", os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, refuse(err)
	}
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
		if err != syscall.EINTR {
			break
		}
	}
	if err != nil {
		f.Close()
		return nil, refuse(err)
	}
	return func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN); _ = f.Close() }, nil
}
