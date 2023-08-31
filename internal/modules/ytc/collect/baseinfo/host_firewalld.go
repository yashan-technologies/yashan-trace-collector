package baseinfo

import (
	"fmt"
	"strings"

	"ytc/defs/bashdef"
	"ytc/defs/runtimedef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/execerutil"
	"ytc/utils/osutil"
	"ytc/utils/userutil"
)

func (b *BaseCollecter) getHostFirewalldStatus() (err error) {
	hostFirewallStatus := datadef.YTCItem{Name: datadef.BASE_HOST_FIREWALLD}
	defer b.fillResult(&hostFirewallStatus)

	log := log.Module.M(datadef.BASE_HOST_FIREWALLD)
	osRelease := runtimedef.GetOSRelease()
	execer := execerutil.NewExecer(log)
	// ubuntu
	if osRelease.Id == osutil.UBUNTU_ID {
		if !userutil.IsCurrentUserRoot() {
			hostFirewallStatus.Error = "checking ubuntu firewall status requires sudo or root"
			hostFirewallStatus.Description = datadef.GenUbuntuFirewalldDesc()
			return
		}
		_, stdout, _ := execer.Exec(bashdef.CMD_BASH, "-c", fmt.Sprintf("%s status", bashdef.CMD_UFW))
		hostFirewallStatus.Details = strings.Contains(stdout, _ubuntu_firewalld_active)
		return
	}
	// other os
	_, stdout, _ := execer.Exec(bashdef.CMD_BASH, "-c", fmt.Sprintf("%s is-active firewalld", bashdef.CMD_SYSTEMCTL))
	hostFirewallStatus.Details = strings.Contains(stdout, _firewalld_active) && !strings.Contains(stdout, _firewalld_inactive)
	return
}
