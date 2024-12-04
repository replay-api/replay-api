terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

provider "aws" {
  region = "us-east-1" # Choose your desired region
}

resource "aws_instance" "game_server" {
  ami           = "ami-0c55b159cbfafe1f0" # Amazon Linux 2 AMI (replace with the appropriate AMI for your region)
  instance_type = "t2.medium"  # Choose an instance type suitable for your server
  # ... (other instance configuration like security groups, tags, etc.)

  user_data = <<-EOF
              #!/bin/bash
              # ... (script to install and run the CS2 server Docker image) ...
              EOF
}

resource "aws_eip" "game_server_eip" {
  instance = aws_instance.game_server.id
  vpc      = true 
}

output "server_address" {
  value = aws_eip.game_server_eip.public_ip
}
