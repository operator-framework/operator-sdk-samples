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
'''


EXAMPLES = '''
'''


import os
import warnings
from re import match


try:
    import MySQLdb
except ImportError:
    mysqldb_found = False
else:
    mysqldb_found = True


from ansible.errors import AnsibleError
from ansible.plugins.lookup import LookupBase
from ansible.module_utils._text import to_native


# the host field will be the terms passed to the lookup
MYSQL_FIELDS = [
   'user', 'passwd', 'db', 'port', 'unix_socket', 'connect_timeout',
   'config_file', 'ssl_cert', 'ssl_key', 'ssl_ca',
]


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


class LookupModule(LookupBase):
    def getstatus(self, cursor, status_name):
        try:
            cursor.execute("SHOW STATUS LIKE %s", (status_name,))
            mysqlstatus_res = cursor.fetchall()
            return mysqlstatus_res
        except:
            # FIXME: proper error handling here
            return None

    def run(self, terms, variables, **kwargs):
        if not mysqldb_found:
            raise AnsibleError("The MySQL-python module is required.")
        else:
            warnings.filterwarnings('error', category=MySQLdb.Warning)

        options = {}
        for field in MYSQL_FIELDS:
            if field in kwargs:
                if field.startswith('ssl_'):
                    if 'ssl' not in options:
                        options['ssl'] = {}
                    ssl_entry = field.replace('ssl_', '')
                    options['ssl'][ssl_entry] = kwargs[field]
                elif field == 'config_file':
                    options['read_default_file'] = kwargs[field]
                else:
                    options[field] = kwargs[field]

        host_statuses = []
        for term in terms:
            host_status = {}
            try:
                options['host'] = term
                db_connection = MySQLdb.connect(**options)
                cursor = db_connection.cursor()
                try:
                    mysqlstatus_res = self.getstatus(cursor, r'%')
                    if mysqlstatus_res:
                        mysqlstatus_res = [(x[0].lower(), typedvalue(x[1])) for x in mysqlstatus_res]
                        for res in mysqlstatus_res:
                            host_status[res[0]] = res[1]
                        host_status["functional"] = True
                    else:
                        raise Exception("No wsrep status variables were found.")
                except Exception as e:
                    host_status["functional"] = False
                    host_status["reason"] = to_native(e)
            except Exception as e:
                host_status["functional"] = False
                if 'read_default_file' in options:
                    if os.path.exists(options['read_default_file']):
                        host_status["reason"] = "Unable to connect to database, check login_user and " \
                                                "login_password are correct or %s has the credentials. " \
                                                "Exception message: %s" % (options['read_default_file'], to_native(e))
                    else:
                        host_status["reason"] = "Unable to find %s. Exception message: %s" % (options['read_default_file'], to_native(e))
                else:
                    host_status["reason"] = "Unable to connect to database on %s: %s" % (term, to_native(e))

            host_statuses.append(host_status)

        return host_statuses

