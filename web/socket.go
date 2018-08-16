package web

import (
	"os"
	"os/user"
	"strconv"

	"github.com/pkg/errors"
)

func (h *Handler) initSocket(addr string) error {
	uid := os.Getuid()
	if h.config.HTTP.SocketUser != "" {
		user, err := user.Lookup(h.config.HTTP.SocketUser)
		if err != nil {
			return errors.Wrap(err, "cannot change socket ownership")
		}
		uid, _ = strconv.Atoi(user.Uid)
	}

	gid := os.Getgid()
	if h.config.HTTP.SocketGroup != "" {
		group, err := user.LookupGroup(h.config.HTTP.SocketGroup)
		if err != nil {
			return errors.Wrap(err, "cannot change socket ownership")
		}
		gid, _ = strconv.Atoi(group.Gid)
	}

	if err := os.Chown(addr, uid, gid); err != nil {
		return errors.Wrap(err, "cannot change socket ownership")
	}

	if h.config.HTTP.SocketMode != "" {
		mode, err := strconv.ParseUint(h.config.HTTP.SocketMode, 8, 32)
		if err != nil {
			return errors.Wrap(err, "cannot change socket permissions")
		}

		err = os.Chmod(addr, os.FileMode(mode))
		if err != nil {
			return errors.Wrap(err, "cannot change socket permissions")
		}
	}

	return nil
}
