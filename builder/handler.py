import os
import sys

from builder import base, builder
from builder.code_check import checker
from builder.code_test import tester
from builder.utils import log


def clean(args):
    if builder.YTCBuilder().clean() != 0:
        return False


def build(args):
    _update_submodules()
    targets = _gen_targets(args)
    if not 'skip_check' in targets:
        if not check(args):
            return False
    if not 'skip_test' in targets:
        if not test(args):
            return False
    if 'force' in targets:
        ret = builder.YTCBuilder().force_build()
    else:
        if 'clean' in targets:
            ret = builder.YTCBuilder().clean()
            if ret != 0:
                return False
        ret = builder.YTCBuilder().build()
    return ret == 0


def check(args):
    passed = checker.Checker().check()
    if not passed:
        log.logger.error('check code failed, please check "code_check.txt" for reason.')
    return passed


def test(args):
    t = tester.Tester()
    log.logger.info('unit test results has been saved to: {}'.format(os.path.join(base.PROJECT_PATH, "unittest")))
    if not t.test():
        return False
    return True


def _update_submodules():
    project_path = base.PROJECT_PATH
    os.chdir(project_path)

    # auto sync the code from git server
    if os.system('git --version') != 0:
        print("command 'git' not found")
        sys.exit(1)

    git_ctrl_path = os.path.join(project_path, ".git")
    if not os.path.exists(git_ctrl_path):
        print("the source code is not under control of git")
        sys.exit(1)

    update_cmd = 'git submodule update --init --recursive'

    ret = os.system(update_cmd)
    if ret != 0:
        sys.exit(1)


def _gen_targets(args):
    running_targets = []
    for arg in dir(args):
        if arg.startswith('_') or arg in ['func', 'jobs']:
            continue
        if not getattr(args, arg, None):
            continue
        running_targets.append(arg)
    return running_targets