# azr
Get resources by Azure ResourceGraph

# Description
`azr` is a tool to list Azure resources.

# Install
1. Go to release page
    - https://github.com/tsubasaxZZZ/azr/releases
2. Download and extract

# Usage
This tool requires Azure CLI and read permission for resources.

## Case 1 : Query from command line option and output to stdout
```bash
# You need login and set target subscription.
az login
az account set --subscription <Your subscriptionID>
# Execute command
./azr  --subscriptionID <Your subscriptionID> -q "resources|where type =~ 'microsoft.compute/virtualmachines'|take 2"
```

## Case 2 : Query from command line option and output to file
```bash
# You need login and set target subscription.
az login
az account set --subscription <Your subscriptionID>
# Execute command
./azr  --subscriptionID <Your subscriptionID> -q "resources|where type =~ 'microsoft.compute/virtualmachines'|take 2" -f result.csv
```
## Case 3 : Query from YAML file
If the first character of the `-q` option starts with `@`, the character string after `@` is regarded as the file path.

```yaml
- name: test1
  query: |
    resources
    | where type =~ "microsoft.compute/virtualmachines"
- name: test2
  query: |
    resources
    | take 10
```

```bash
# You need login and set target subscription.
az login
az account set --subscription <Your subscriptionID>
# Execute command
./azr  --subscriptionID <Your subscriptionID> -q @sample/query.yml
```

Result is output file to :`<name>.csv`

## Help
```bash
NAME:
   azr - Azure Resource Graph Command

USAGE:
   azr [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --subscriptionID value
   --query value, -q value
   --file value, -f value   Speify output filepath(If not specify, out to stdout)
   --help, -h               show help (default: false)
```

# Sample
Please see sample directory.
