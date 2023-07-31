import os
from typing import List

from builder import base
from builder.code_check import reporter
from builder.utils import execer
from builder.utils import log

_REQUIRES: List[str] = [
    'golangci-lint',
    'mypy',
    'yapf',
    'shellcheck',
]

_GO_FILES: List[str] = [
    base.PROJECT_PATH,
]

_PYTHON_FILES: List[str] = [
    os.path.join(base.PROJECT_PATH, 'builder'),
]

_SHELL_FILES: List[str] = [
    os.path.join(base.PROJECT_PATH, 'scripts', '*.sh'),
]


class Checker(reporter.Reporter):

    def __init__(self):
        super(Checker, self).__init__()
        self._requires = _REQUIRES

    def check(self) -> bool:
        if not self._prepare():
            return False
        log.logger.info('checking code starting...')
        is_code_format_valid = CodeFormatChecker().check()
        is_code_lint_valid = CodeLintChecker().check()
        passed = is_code_format_valid and is_code_lint_valid
        if not passed:
            self._write_report('Abort building...\n')
        log.logger.info('checking code finished.')
        log.logger.info('result has been saved to: {}'.format(self._report_file))
        return passed

    def _prepare(self) -> bool:
        if os.path.exists(self._get_report()):
            os.remove(self._get_report())
        log.logger.info('checking requires...')
        success = True
        for r in self._requires:
            ret, _, _ = execer.exec('command -v {}'.format(r))
            if ret != 0:
                success = False
                self._write_report('{}: command not found'.format(r))
        return success


class CodeFormatChecker(reporter.Reporter):

    def __init__(self, ):
        super(CodeFormatChecker, self).__init__()

    def check(self):
        self._write_report('Check code format starting...\n')
        python_format_passed = self._check_python_format()
        passed = python_format_passed
        if passed:
            self._write_report('Check code format passed.\n')
        return passed

    # yapf
    def _check_python_format(self):
        conf = os.path.join(base.PROJECT_PATH, '.style.yapf')
        cmd = 'yapf -rdp --style {}'.format(conf)
        for f in _PYTHON_FILES:
            cmd += ' {}'.format(f)
        self._write_report(cmd + '\n')
        ret, out, _ = execer.exec(cmd)
        self._write_report(out)
        passed = ret == 0
        if not passed:
            self._write_report('Fail to pass yapf.\n')
        else:
            self._write_report('Succeed to pass yapf.\n')
        return passed


class CodeLintChecker(reporter.Reporter):

    def __init__(self, ):
        super(CodeLintChecker, self).__init__()

    def check(self):
        self._write_report('Check code lint starting...\n')
        go_lint_passed = self._check_go_lint()
        python_lint_passed = self._check_python_lint()
        # shell_lint_passed = _shell_lint()
        # TODO: open shell lint when scripts is not empty
        shell_lint_passed = True
        passed = go_lint_passed and python_lint_passed and shell_lint_passed
        if passed:
            self._write_report('Check code lint passed.\n')
        return passed

    # golangci-lint
    def _check_go_lint(self):
        conf = os.path.join(base.PROJECT_PATH, '.golangci.yml')
        ret_list = []
        for f in _GO_FILES:
            cmd = 'cd {};golangci-lint run --config {}'.format(f, conf)
            self._write_report(cmd + '\n')
            ret, out, err = execer.exec(cmd)
            passed = ret == 0
            self._write_report(out)
            if not passed:
                self._write_report(err)
            ret_list.append(passed)
        passed = True
        for ret in ret_list:
            if not ret:
                passed = False
        if not passed:
            self._write_report('Fail to pass golangci-lint.\n')
        else:
            self._write_report('Succeed to pass golangci-lint.\n')
        return passed

    # mypy
    def _check_python_lint(self):
        conf = os.path.join(base.PROJECT_PATH, 'mypy.ini')
        cmd = 'mypy --config {}'.format(conf)
        for f in _PYTHON_FILES:
            cmd += ' {}'.format(f)
        self._write_report(cmd + '\n')
        ret, out, _ = execer.exec(cmd)
        self._write_report(out)
        passed = ret == 0
        if not passed:
            self._write_report('Fail to pass mypy.\n')
        else:
            self._write_report('Succeed to pass mypy.\n')
        return passed

    # shellcheck
    def _check_shell_lint(self):
        cmd = 'shellcheck'
        for f in _SHELL_FILES:
            cmd += ' {}'.format(f)
        self._write_report(cmd + '\n')
        ret, out, _ = execer.exec(cmd)
        self._write_report(bytes.decode(out))
        passed = ret == 0
        if not passed:
            self._write_report('Fail to pass shellcheck.\n')
        else:
            self._write_report('Succeed to pass shellcheck.\n')
        return passed
