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
        ret = builder.YTCBuilder().build()
    if ret != 0:
        return False


def check(args):
    passed = checker.Checker().check()
    if not passed:
        log.logger.error('check code failed, please check "code_check.txt" for reason.')
    return passed


def test(args):
    return tester.code_test()


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