#!/usr/bin/python
# -*- coding: utf-8 -*-

# (c) 2018, James Cammarata <jimi@sngx.net>
# Copied and modified mainly from the mysql_variables.py module.
#
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}


DOCUMENTATION = '''
---
module: mysql_status

short_description: Get MySQL status variables.
description:
    - Query MySQL status variables
version_added: 2.8
author: "James Cammarata"
options:
    status:
        description:
            - Variable name to operate
        required: True
extends_documentation_fragment: mysql
'''
EXAMPLES = '''
# Get Galera cluster size
- mysql_status:
    status: wsrep_cluster_size

# Get all wsrep status variables
- mysql_status:
    status: "%wsrep%"
'''

import json
import os
import warnings
from re import match

try:
    import MySQLdb
except ImportError:
    mysqldb_found = False
else:
    mysqldb_found = True

from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils.database import SQLParseError, mysql_quote_identifier
from ansible.module_utils.mysql import mysql_connect, mysqldb_found
from ansible.module_utils._text import to_native


def typedvalue(value):
    """
    Convert value to number whenever possible, return same value
    otherwise.

    >>> typedvalue('3')
    3
    >>> typedvalue('3.0')
    3.0
    >>> typedvalue('foobar')
    'foobar'

    """
    try:
        return int(value)
    except ValueError:
        pass

    try:
        return float(value)
    except ValueError:
        pass

    return value


def getstatus(cursor, status_name):
    if 1: #try:
        cursor.execute("SHOW STATUS LIKE %s", (status_name,))
        mysqlstatus_res = cursor.fetchall()
        return mysqlstatus_res
    #except:
    #    # FIXME: proper error handling here
    #    return None


def main():
    module = AnsibleModule(
        argument_spec=dict(
            status=dict(default=None, type='list', required=True),
            login_user=dict(default=None),
            login_password=dict(default=None, no_log=True),
            login_host=dict(default="localhost"),
            login_port=dict(default=3306, type='int'),
            login_unix_socket=dict(default=None),
            ssl_cert=dict(default=None),
            ssl_key=dict(default=None),
            ssl_ca=dict(default=None),
            connect_timeout=dict(default=30, type='int'),
            config_file=dict(default="~/.my.cnf", type="path")
        )
    )
    user = module.params["login_user"]
    password = module.params["login_password"]
    ssl_cert = module.params["ssl_cert"]
    ssl_key = module.params["ssl_key"]
    ssl_ca = module.params["ssl_ca"]
    connect_timeout = module.params['connect_timeout']
    config_file = module.params['config_file']
    db = 'mysql'

    mysqlstatus = module.params["status"]
    if not mysqldb_found:
        module.fail_json(msg="The MySQL-python module is required.")
    else:
        warnings.filterwarnings('error', category=MySQLdb.Warning)

    try:
        cursor = mysql_connect(
            module,
            user,
            password,
            config_file,
            ssl_cert,
            ssl_key,
            ssl_ca,
            connect_timeout=connect_timeout,
        )
    except Exception as e:
        if os.path.exists(config_file):
            module.fail_json(msg="unable to connect to database, check login_user and login_password are correct or %s has the credentials. "
                                 "Exception message: %s" % (config_file, to_native(e)))
        else:
            module.fail_json(msg="unable to find %s. Exception message: %s" % (config_file, to_native(e)))

    statuses = {}
    for status in mysqlstatus:
        if match('^[0-9a-z_\%]+$', status) is None:
            module.fail_json(msg="invalid status name \"%s\"" % status)
        mysqlstatus_res = getstatus(cursor, status)
        if mysqlstatus_res is None:
            statuses[status] = None
        else:
            mysqlstatus_res = [(x[0].lower(), typedvalue(x[1])) for x in mysqlstatus_res]
            for res in mysqlstatus_res:
                statuses[res[0]] = res[1]

    module.exit_json(msg="Status found", status=statuses, changed=False)


if __name__ == '__main__':
    main()
