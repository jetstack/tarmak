resource "aws_iam_role" "jenkins" {
  name               = "${data.template_file.stack_name.rendered}-jenkins"
  path               = "/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "jenkins" {
  name = "${data.template_file.stack_name.rendered}-jenkins"
  role = "${aws_iam_role.jenkins.name}"
}

# TODO: Warning this grants all (!!) rights to the jenkins instance
resource "aws_iam_policy" "jenkins" {
  name   = "${data.template_file.stack_name.rendered}-jenkins"
  path   = "/"
  policy = "${file("${path.module}/templates/jenkins_role_policy.json")}"
}

resource "aws_iam_policy_attachment" "jenkins" {
  name       = "${aws_iam_policy.jenkins.name}"
  roles      = ["${aws_iam_role.jenkins.name}"]
  policy_arn = "${aws_iam_policy.jenkins.arn}"
}
