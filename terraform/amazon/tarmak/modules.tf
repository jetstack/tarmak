module "network" {
  source = "../modules/network"

  network = "${var.network}"
  name = "${var.name}"
  project = "${var.project}"
  contact = "${var.contact}"
  region = "${var.region}"
  peer_vpc_id = "${var.peer_vpc_id}"
  availability_zones = "${var.availability_zones}"
  stack = "${var.stack}"
  state_bucket = "${var.state_bucket}"
  stack_name_prefix = "${var.stack_name_prefix}"
  allowed_account_ids = "${var.allowed_account_ids}"
  vpc_peer_stack = "${var.vpc_peer_stack}"
  environment = "${var.environment}"
  private_zone = "${var.private_zone}"
  state_cluster_name = "${var.state_cluster_name}"
  vpc_net = "${var.vpc_net}"
  route_table_public_ids = "${var.route_table_public_ids}"
  route_table_private_ids = "${var.route_table_private_ids}"
  private_zone_id = "${var.private_zone_id}"
}

module "bastion" {
  source = "../modules/bastion"

  public_zone = "${var.public_zone}"
  environment = "${var.environment}"
  stack_name_prefix = "${var.stack_name_prefix}"
  name = "${var.name}"
  vpc_id = "${module.network.vpc_id}"
  project = "${var.project}"
  contact = "${var.contact}"
  bastion_ami = "${var.bastion_ami}"
  bastion_instance_type = "${var.bastion_instance_type}"
  public_subnet_ids = "${module.network.public_subnet_ids}"
  key_name = "${var.key_name}"
  bastion_root_size = "${var.bastion_root_size}"
  admin_ips = "${var.admin_ips}"
  public_zone_id = "${var.public_zone_id}"
  private_zone_id = "${module.network.private_zone_id[0]}"
}

/*module "bastion" {
  source  = "../modules/bastion"
  
  servers = 3
}*/