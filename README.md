# glueprint

`glueprint` is a demo project covering the basics of configuration management.

*Disclaimer:* This is a demonstration. There are several liberties taken for convenience.

## Architecture

The application is written and compiled in Go.

For the purposes of this demonstration, this tool works against an Ubuntu instance. It could be expanded to factor in state management and manage other resources.

## Installation

TODO

## Configuration

All configuration for managed resources must be specified in a file called `glue.yaml`. The file is written with `yaml` and structured using the following elements.

### Host & Password

The IP address of the host to be managed and the corresponding password.

*Disclaimer:* The demonstration leverages plaintext connections. In a real-world scenario, you would use appropriate authentication.

### Files

Adding a file to this list will create it on the managed host. Removing it will delete the file.

Changing any of the fields or the content of the file will cause an update to the file on the host.

The `name` field in each item looks for a file in the directory in which the tool is being invoked. For the sake of this demo, all files should be in the same directory as the `glue.yaml` file.

`path` specifies the directory in which the file will be created on the host.

`mode` describes the permissions applied to the file.

```yaml
files:
  - name: index.php
    path: /var/www/html
    mode: 0600
  - name: dir.conf
    path: /etc/apache2/mods-enabled
    mode: 0600
```

### Packages

Adding a package to this list will intall it on the managed host.

```yaml
packages:
  - package: apache2
  - package: php
    version: something
```

### Command

```yaml
command: ['service', 'apache2', 'restart']
```

A command provided here acts as an after-deploy hook. It will be run at the end of the deploy.

### Full Example

```yaml
webserver:
  host: 1.2.3.4
  password: foo
  files:
    - name: index.php
      path: /var/www/html
      mode: 0600
    - name: dir.conf
      path: /etc/apache2/mods-enabled
      mode: 0600
  packages:
    - package: apache2
    - package: php
  command: ['service', 'apache2', 'restart']
somethingelse:
  host: 5.6.7.8
  password: bar
  files:
    - name: index.php
      path: /var/www/html
      mode: 0600
    - name: dir.conf
      path: /etc/apache2/mods-enabled
      mode: 0600
  packages:
    - package: apache2
    - package: php
  command: ['service', 'apache2', 'restart']
```

## Usage

After a configuration file has been created, the following commands can be leveraged:

### `glueprint propose`

This will show proposed changes based on the requested configuration.

### `glueprint deploy`

This will deploy changes to the managed resource.

A state file called `glueprint-state.json` will be created to manage resources.

## Opportunities

- At least one file and one package should be specified for the demonstration.
- It is never acceptable to put a password in the config file, but it's easiest for this demonstration.
- Since this method uses root creds, package and file manipulation doesn't depend on `sudo` - in a proper rollout, this would be handled more securely.
- Package manipulation depends on `apt`. Any requested file should be available in the standard repository.
- This doesn't verify connectivity to the host nor does it add the `ssh` key to known hosts. You will need to connect manually to the host at least once.
