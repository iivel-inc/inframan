# Writing Terranix Infrastructure Configs

Infrastructure in inframan is defined using [Terranix](https://terranix.org/) -- Nix attribute sets that compile to Terraform JSON. This gives you the full power of the Nix language (functions, imports, conditionals) while targeting Terraform's resource model.

## Basics

A Terranix config is a Nix expression where attribute paths map to Terraform HCL concepts:

| Nix Attribute | Terraform Equivalent |
|---------------|---------------------|
| `provider.aws` | `provider "aws" { }` |
| `resource.aws_instance.main` | `resource "aws_instance" "main" { }` |
| `data.aws_ami.nixos` | `data "aws_ami" "nixos" { }` |
| `variable.instance_type` | `variable "instance_type" { }` |
| `output.public_ip` | `output "public_ip" { }` |

## Required Outputs

Inframan reads Terraform outputs to discover instance IPs. Your infrastructure config **must** include at least one of these output formats:

### Single Instance (Legacy)

```nix
output.public_ip = {
  value = "\${aws_instance.main.public_ip}";
};

# Optional: for private networking
output.private_ip = {
  value = "\${aws_instance.main.private_ip}";
};
```

### Multiple Instances

For projects with multiple instances, use an `instances` map:

```nix
output.instances = {
  value = {
    "web-1" = "\${aws_instance.web[0].public_ip}";
    "web-2" = "\${aws_instance.web[1].public_ip}";
    "db-1"  = "\${aws_instance.db.public_ip}";
  };
};
```

This enables `inframan ssh project/web-1` syntax.

## Variable Interpolation

In Nix strings, `${}` is Nix interpolation. To produce Terraform `${...}` references, escape the dollar sign:

```nix
# Correct: produces ${var.instance_type} in JSON
instance_type = "\${var.instance_type}";

# Wrong: Nix tries to evaluate var.instance_type as a Nix variable
instance_type = "${var.instance_type}";
```

## AWS Patterns

### Provider Configuration

```nix
provider.aws = {
  region = "us-east-1";
  access_key = "\${var.aws_access_key}";
  secret_key = "\${var.aws_secret_key}";
};
```

### NixOS AMI Data Source

NixOS publishes official AMIs under owner `427812963091`:

```nix
data.aws_ami.nixos = {
  most_recent = true;
  owners = [ "427812963091" ];
  filter = [
    { name = "name"; values = [ "nixos/24.05*" ]; }
    { name = "architecture"; values = [ "x86_64" ]; }
  ];
};
```

### Security Group

```nix
resource.aws_security_group.main = {
  name = "my-sg";
  description = "Inframan-managed security group";

  ingress = [
    {
      description = "SSH";
      from_port = 22;
      to_port = 22;
      protocol = "tcp";
      cidr_blocks = [ "0.0.0.0/0" ];
      # All these fields are required to avoid Terraform diff issues:
      ipv6_cidr_blocks = [];
      prefix_list_ids = [];
      security_groups = [];
      self = false;
    }
  ];

  egress = [
    {
      description = "All outbound";
      from_port = 0;
      to_port = 0;
      protocol = "-1";
      cidr_blocks = [ "0.0.0.0/0" ];
      ipv6_cidr_blocks = [];
      prefix_list_ids = [];
      security_groups = [];
      self = false;
    }
  ];
};
```

### EC2 Instance

```nix
resource.aws_instance.main = {
  ami = "\${data.aws_ami.nixos.id}";
  instance_type = "\${var.instance_type}";
  key_name = "\${aws_key_pair.deployer.key_name}";
  vpc_security_group_ids = [ "\${aws_security_group.main.id}" ];

  root_block_device = {
    volume_size = 20;
    volume_type = "gp3";
  };

  tags = {
    Name = "my-instance";
    ManagedBy = "inframan";
  };
};
```

### Key Pair

```nix
variable.ssh_public_key = {
  type = "string";
};

resource.aws_key_pair.deployer = {
  key_name = "inframan-deployer";
  public_key = "\${var.ssh_public_key}";
};
```

## Remote Backend

To use S3 for Terraform state storage instead of local state:

```nix
terraform.backend.s3 = {
  bucket = "my-terraform-state";
  key = "inframan/production/terraform.tfstate";
  region = "us-east-1";
  encrypt = true;
  dynamodb_table = "terraform-locks";
};
```

With a remote backend, `inframan init` becomes essential for pulling state on new machines before running `deploy` or `ssh`.

## Variables

Define Terraform variables with types, defaults, and descriptions:

```nix
variable.instance_type = {
  description = "EC2 instance type";
  type = "string";
  default = "t3.micro";
};

variable.aws_access_key = {
  description = "AWS Access Key";
  type = "string";
  sensitive = true;
};
```

Variables are provided via `terraform.tfvars` files or `TF_VAR_*` environment variables. See [Configuration](configuration.md) for details.

## Using Nix Features

Since this is Nix, you can use the full language:

```nix
let
  region = "us-east-1";
  commonTags = {
    ManagedBy = "inframan";
    Environment = "production";
  };
in
{
  provider.aws = { inherit region; };

  resource.aws_instance.main = {
    # ...
    tags = commonTags // { Name = "web-server"; };
  };

  resource.aws_instance.db = {
    # ...
    tags = commonTags // { Name = "database"; };
  };
}
```

You can also use imports to share common configuration across projects:

```nix
# common.nix
{ commonTags, region }:
{
  provider.aws = { inherit region; };
  # shared resources...
}

# infrastructure.nix
let
  common = import ./common.nix {
    commonTags = { ManagedBy = "inframan"; };
    region = "us-east-1";
  };
in
common // {
  # project-specific resources...
}
```

## Complete Example

See [example/infrastructure.nix](../example/infrastructure.nix) for a complete working example with AWS provider, NixOS AMI, security group, EC2 instance, and outputs.
