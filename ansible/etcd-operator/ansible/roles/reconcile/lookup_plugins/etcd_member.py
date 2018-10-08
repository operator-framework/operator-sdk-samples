# python 3 headers, required if submitting to Ansible
from __future__ import (absolute_import, division, print_function)

__metaclass__ = type

DOCUMENTATION = """
      lookup: etcd_member
        author: Alay Patel <alay1431@gmail.com>
        version_added: "0.1"
        short_description: look up members in etcd cluster
        description:
            - This lookup returns the members etcd cluster
        options:
         _cluster_host:
            description: reachable ip of the etcd cluster
            required: True
         _cluster_port:
            description: port that etcd cluster is listening on
            required: True
        notes:
        
"""
from ansible.errors import AnsibleError, AnsibleParserError
from ansible.plugins.lookup import LookupBase

try:
    from __main__ import display
except ImportError:
    from ansible.utils.display import Display

    display = Display()

import etcd3


class LookupModule(LookupBase):
    def run(self, terms, **kwargs):
        # lookups in general are expected to both take a list as input and output a list
        # this is done so they work with the looping construct 'with_'.
        kwargs.setdefault('cert_cert', None)
        kwargs.setdefault('ca_cert', None)
        kwargs.setdefault('cert_key', None)
        client = etcd3.client(host=kwargs['cluster_host'], port=kwargs['cluster_port'],
                              cert_cert=kwargs['cert_cert'],
                              cert_key=kwargs['cert_key'],
                              ca_cert=kwargs['ca_cert'])
        ret = [dict(id=m.id, name=m.name, peer_urls=m.peer_urls, client_urls=m.client_urls)
               for m in client.members]

        return ret
