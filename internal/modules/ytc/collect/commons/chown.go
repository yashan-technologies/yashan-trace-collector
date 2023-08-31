package ytccollectcommons

import (
	"os"
	"strconv"

	"ytc/defs/runtimedef"
	"ytc/utils/userutil"
)

func ChownToExecuter(path string) error {
	if !userutil.IsCurrentUserRoot() {
		return nil
	}
	user := runtimedef.GetExecuter()
	uid, _ := strconv.ParseInt(user.Uid, 10, 64)
	gid, _ := strconv.ParseInt(user.Gid, 10, 64)
	if uid == userutil.ROOT_USER_UID {
		return nil
	}
	return os.Chown(path, int(uid), int(gid))
}
