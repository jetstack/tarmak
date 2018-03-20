data "aws_route_table" "public" {
  count     = "${length(data.aws_subnet.public.*.id)}"
  subnet_id = "${data.aws_subnet.public.*.id[count.index]}"
}

data "aws_route_table" "private" {
  count     = "${length(data.aws_subnet.private.*.id)}"
  subnet_id = "${data.aws_subnet.private.*.id[count.index]}"
}
