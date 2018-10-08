#!/usr/bin/python

ANSIBLE_METADATA = {
    'metadata_version': '1.1',
    'status': ['preview'],
    'supported_by': 'community'
}

DOCUMENTATION = '''
---
module: etcd_cluster

short_description: This is my etcd_cluster module

version_added: "2.7"

description:
    - "This is my longer description explaining my etcd_cluster module"

options:
    state:
        description:
            - `present` adds the member to the cluster, `absent` removes the member from cluster
        required: true
    name:
        description:
            - name of the member required if the state is `present`
        required: false
    id:
        description:
            - id of the member, required if the state is `absent`
        required: false

    
    cluster_host:
        description:
            - reachable host on which cluster is listening
        required: true

    cluster_port:
        description:
            - port on which the cluster is listening on
        required: false

requirements
    - etcd3

author:
    - Alay Patel(@alaypatel07)
'''

EXAMPLES = '''
# Pass in a message
- name: Add a member to etcd cluster reachable at 192.168.39.66:32379
  etcd:
    name: hello world
    state: "present",
    cluster_host: "192.168.39.66",
    cluster_port: "32379",
    name: "hello-world",
    peer_urls: 
      - "http://hello-world.default.svc:2380"
    


# pass in a message and have changed true
- name: Remove a member from etcd cluster reachable at 192.168.39.66:32379
  etcd:
    state: "absent",
    cluster_host: "192.168.39.66",
    cluster_port: "32379",
    id: "14181629488891917781"

'''

RETURN = '''
original_message:
    description: The original name param that was passed in
    type: str
message:
    description: The output message that the sample module generates
'''

from ansible.module_utils.basic import AnsibleModule
import etcd3


def run_module():
    # define the available arguments/parameters that a user can pass to
    # the module
    module_args = dict(
        state=dict(type='str', required=True),
        name=dict(type='str', required=False, default=''),
        id=dict(type='str', required=False, default=''),
        ca_cert=dict(type='str', required=False, default=None),
        cert_key=dict(type='str', required=False, default=None),
        cert_cert=dict(type='str', required=False, default=None),
        cluster_host=dict(type='str', required=True),
        cluster_port=dict(type='str', required=False, default=''),
        peer_urls=dict(type='list', required=False, default=[])
    )

    result = dict(
        members=[]
    )

    # the AnsibleModule object will be our abstraction working with Ansible
    # this includes instantiation, a couple of common attr would be the
    # args/params passed to the execution, as well as if the module
    # supports check mode
    module = AnsibleModule(
        argument_spec=module_args,
        supports_check_mode=True
    )

    # if the user is working with this module in only check mode we do not
    # want to make any changes to the environment, just return the current
    # state with no modifications
    if module.check_mode:
        return result

    if module.params['state'] == 'present':
        try:
            members = etcd_member_present(cluster_host=module.params['cluster_host'],
                                          cluster_port=module.params['cluster_port'],
                                          name=module.params['name'],
                                          peer_urls=module.params['peer_urls'],
                                          ca_cert=module.params['ca_cert'],
                                          cert_cert=module.params['cert_cert'],
                                          cert_key=module.params['cert_key'])
            result['members'] = [
                dict(id=m.id, name=m.name, peer_urls=list(m.peer_urls), client_urls=list(m.client_urls))
                for m in members]
        except Exception as err:
            if str(err) == 'name':
                module.fail_json(msg='name is empty', **result)
            if str(err) == 'peer_urls':
                module.fail_json(msg='peer_urls is empty', **result)

    elif module.params['state'] == 'absent':
        try:
            members = etcd_member_absent(cluster_host=module.params['cluster_host'],
                                         cluster_port=module.params['cluster_port'],
                                         id=module.params['id'],
                                         ca_cert=module.params['ca_cert'],
                                         cert_cert=module.params['cert_cert'],
                                         cert_key=module.params['cert_key']
                                         )
            result['members'] = [
                dict(id=m.id, name=m.name, peer_urls=list(m.peer_urls), client_urls=list(m.client_urls))
                for m in members]
        except ValueError as err:
            module.fail_json(msg=str(err), **result)
        except Exception as err:
            if str(err) == 'id':
                module.fail_json(msg='id is not set', **result)
    module.exit_json(**result)


def etcd_member_present(**kwargs):
    if kwargs['name'] == '':
        raise Exception('name')
    if len(kwargs['peer_urls']) == 0:
        raise Exception('peer_urls')

    client = connect_etcd_cluster(kwargs['cluster_host'], kwargs['cluster_port'], kwargs['ca_cert'], kwargs['cert_key'],
                                  kwargs['cert_cert'])
    add_member(client, kwargs['peer_urls'])
    return client.members


def etcd_member_absent(**kwargs):
    if kwargs['id'] == -1:
        raise Exception('id')
    client = connect_etcd_cluster(kwargs['cluster_host'], kwargs['cluster_port'], kwargs['ca_cert'], kwargs['cert_key'],
                                  kwargs['cert_cert'])
    if kwargs['id'][0:2] == '0x':
        _id = int(kwargs['id'], 16)
    else:
        _id = int(kwargs['id'])
    remove_member(client, _id)
    return client.members


def connect_etcd_cluster(host, port, ca_cert=None, cert_key=None, cert_cert=None):
    return etcd3.client(host=host, port=port, ca_cert=ca_cert, cert_key=cert_key, cert_cert=cert_cert)


def add_member(client, urls):
    client.add_member(urls=urls)


def remove_member(client, id):
    client.remove_member(id)


def main():
    run_module()


if __name__ == '__main__':
    main()
