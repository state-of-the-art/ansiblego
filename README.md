# AnsibleGo

AnsibleGo is an image configuration tool. It was born as a rewrite of Ansible using golang due to:

* Configuration-management is obviosuly too much complicated for simple image building
* Too much dependencies - python sometimes is too much for docker image and even for big images
* Non-unified commands infrastructure - `win_` prefix is not a good way to unify the logic

AnsibleGo wants you to write simple playbooks and roles and not overcomplicate the logic.

The architecture is local execution with an ability to push and run remote agent to execute the
specified commands.

## Features

To address the main Ansible pain-points the AnsibleGo have the next features:

* One executable binary
   > Simple to install by just copying to the target system.
* No external dependencies
   > Especially important for the minimal environments. In Ansible that could be partially solved
   > by standalone python but pip again is needed makes a larger footprint which is not ideal.
* Built-in support for SSH and WinRM transports
   > No need to install additional dependencies - it works out of the box to cover the majority of
   > the potential targets.
* Built-in SSHD for compact systems
   > In order to get access to the system you can just push the AnsibleGo to target system and run
   > it in agent mode to get access to remote system from host system. Especially useful for docker
   > and mobile platforms.
* Almost identical dialect of Ansible playbooks aimed to simplicity
   > No need to learn it again - just write the playbooks and roles the way you know it.
* Supports scripted modules to simply extend the functionality
   > Just like in Ansible you are able to add or override module part of the system to quickly
   > add or improve and not to suffer much from the bug you found in your particular case.
* Aimed to images building - towards installing, not the configuration management
   > Taking into account the complex configuration management systems which tries to manage all the
   > possible previous states of the environment and move them to the known state (nonsence) this
   > project encourages to simplify the logic by moving towards image management in modern infra.

## Goal

The project goal is to replace Ansible in image building configuration management operations using
the known 2.9 playbooks specification interface and simplify the run experience.

Potentially the future steps will be aimed to prepare a dialect to define the playbook/roles specs
for simplicity based on the received experience with implementing the Ansible playbook structure.

## Initial PoC functionality

* Multiarch executable
   * The build script is separated and just packs/combines the built binaries
   * Supports GZ, XZ and UPX (which have it's own issues) packing for binaries to reduce exec size
   * It seems combined execs are working well on Linux, MacOS and Windows
   * Unix executable based on sh script for Mac/Linux hosts
   * Prepare the extractor of the needed arch from binary
* Modules plugins via scripting
   * Can embed them
   * Can run various functions with native interface
   * Performance is good enough
* Parsing of the simple ansible playbooks/roles with templates
* SSH/WinRM remote client support
* TODO: Minimal SSHD transport for the agent mode
* TODO: Builtin `SO_DONTROUTE` local proxy

### WinRM remote setup

It's a good idea to use https connection for winrm, especially if it requires just a couple of
additional comamnds:

1. Create certificate:
   ```
   $cert = New-SelfSignedCertificate -Subject 'CN=winrm-server' -TextExtension '2.5.29.37={text}1.3.6.1.5.5.7.3.1'
   $tp = $cert.Thumbprint
   ```
2. Create winrm listener (use thumbprint from previous command output):
   ```
   winrm create winrm/config/Listener?Address=*+Transport=HTTPS '@{Hostname="winrm-server"; CertificateThumbprint="<cert thumbprint here>"}'
   ```
3. Run winrm quickconfig with https:
   ```
   winrm quickconfig -transport:https
   winrm set winrm/config/service/Auth '@{Basic="true"}'
   ```
4. Allow firewall rule:
   ```
   $FirewallParam = @{DisplayName='WinRM (HTTPS-In)' Direction='Inbound' LocalPort=5986 Protocol='TCP' Action='Allow' Program='System'}
   New-NetFirewallRule @FirewallParam
   ```
