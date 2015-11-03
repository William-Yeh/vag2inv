vag2inv
==============

[![Circle CI](https://circleci.com/gh/William-Yeh/vag2inv.svg?style=shield)](https://circleci.com/gh/William-Yeh/vag2inv) [![Build Status](https://travis-ci.org/William-Yeh/vag2inv.svg?branch=master)](https://travis-ci.org/William-Yeh/vag2inv)


This program generates Ansible inventory file by investigating status from a set of running Vagrant boxes.

It was initially invented as an auxiliary tool for running Ansible + Vagrant under the Windows host machine. Under such circumstances, Ansible is not natively supported; the Vagrant's "[Ansible provisioner mechanism](https://docs.vagrantup.com/v2/provisioning/ansible.html)" is unavailable, let alone the `.vagrant/provisioners/ansible/inventory/vagrant_ansible_inventory` file. This program helps generate appropriate inventory file for Ansible to use.

Of course, this program is not limited to Windows users. Users of other platforms still benefits from this program if they want an Ansible-centric workflow, and dislike the lenghy path `.vagrant/provisioners/ansible/inventory/vagrant_ansible_inventory`.



## Installation

1. Browse the [releases](https://github.com/William-Yeh/vag2inv/releases) page.

2. Choose the specific platform, architecture, and version of the executable file.

3. Download the executable file to any place in your `PATH`.

4. Rename it to `vag2inv.exe` (Windows) or `vag2inv` (Linux and Mac OS X) for your convenience.



## Usage


```
Generate Ansible inventory file from Vagrant.

Usage:
  vag2inv  [options]  <inventory_filename>
  vag2inv  --help
  vag2inv  --version

Options:
  -d, --stdout                Also dump to stdout.
  -f, --force                 Force overwrite inventory file;
                                [default: false].
  --vm                        Compatible for Ansible control machine that resides in VM;
                                [default: false].
  -p <dir>, --prefix <dir>    Rewrite the prefix part of the private key's path.

```



## Guide for non-Windows users

For platforms *with* native Ansible support: Linux, Mac OS X, etc.


#### Single managed node

In the `examples/single` directory:

1. Start the Vagrant box (`ubuntu/trusty64`):

   ```
   vagrant up
   ```

2. Generate inventory file (with the name `HOSTS-VAGRANT`):

   ```
   vag2inv HOSTS-VAGRANT
   ```

3. Use `ansible` to see host information of this box:

   ```
   ansible -i HOSTS-VAGRANT all -m setup
   ```

4. Use `ansible-playbook` to apply Ansible playbook to this box:

   ```
   ansible-playbook -i HOSTS-VAGRANT playbook.yml
   ```



#### Multiple managed nodes

In the `examples/multiple` directory:

1. Start the Vagrant boxes (`ubuntu/trusty64` and `bento/centos-7.1`):

   ```
   vagrant up
   ```

2. Generate inventory file (with the name `HOSTS-VAGRANT`):

   ```
   vag2inv HOSTS-VAGRANT
   ```

3. Use `ansible` to see host information of these boxes:

   ```
   ansible -i HOSTS-VAGRANT all -m setup
   ```

4. Use `ansible-playbook` to apply Ansible playbook to these boxes:

   ```
   ansible-playbook -i HOSTS-VAGRANT playbook.yml
   ```



## Guide for all platform users

For platforms *with* or *without* native Ansible support, including Windows.

This guide demonstrates how to use a Ansible-in-VM as the Ansible control machine (refer to the "[Vagrant Box for Ansible Control Machine](https://github.com/William-Yeh/ansible-vagrantbox)" project for more information).


#### Single managed node

In the `examples/single` directory:

1. Start the Vagrant box (`ubuntu/trusty64`):

   ```
   vagrant up
   ```

2. Generate inventory file (this time, with the `--vm` option):

   ```
   vag2inv --vm -f HOSTS-VAGRANT
   ```

#### Multiple managed nodees

In the `examples/multiple` directory:

1. Start the Vagrant boxes (`ubuntu/trusty64` and `bento/centos-7.1`):

   ```
   vagrant up
   ```

2. Generate inventory file (this time, with the `--vm` option):

   ```
   vag2inv --vm -f HOSTS-VAGRANT
   ```

#### Control machine

In the `examples/control-machine` directory:

1. Start and login to the Vagrant box (`williamyeh/ansible`):

   ```
   vagrant up
   vagrant ssh
   ```


2. (Optionally) uncomment the `export ANSIBLE_HOST_KEY_CHECKING=false`
 line near the end of `/home/vagrant/.zshrc` to disable the "[Host Key Checking](http://docs.ansible.com/ansible/intro_getting_started.html#host-key-checking)" feature.

3. Use `ansible` to see host information of these boxes:

   ```
   cd /vagrant/examples/single
   ansible -i HOSTS-VAGRANT all -m setup

   cd /vagrant/examples/multiple
   ansible -i HOSTS-VAGRANT all -m setup
   ```

4. Use `ansible-playbook` to apply Ansible playbook to these boxes:

   ```
   cd /vagrant/examples/single
   ansible-playbook -i HOSTS-VAGRANT playbook.yml

   cd /vagrant/examples/multiple
   ansible-playbook -i HOSTS-VAGRANT playbook.yml
   ```



## Build, if you want...

Build the executable for your platform (before compiling, please make sure that you have [Go](https://golang.org/) compiler installed):

```
$ ./build.sh
```

Or, build the executables with [Docker Compose](https://docs.docker.com/compose/):

```
$ docker-compose up
```

Or, build the executables with [Vagrant](https://www.vagrantup.com/):

```
$ vagrant up
```

It will place the `vag2inv-i386.exe`, `vag2inv-x86_64.exe`, etc. executables into the `dist` directory.




## History

- 0.1 - Initial release.


## Author

William Yeh, william.pjyeh@gmail.com

## License

Apache License V2.0.  See [LICENSE](LICENSE) file for details.
