# TODO: Warning this grants all (!!) rights to the jenkins instance
resource "aws_iam_policy" "jenkins" {
  name   = "${data.template_file.stack_name.rendered}-jenkins-custom"
  path   = "/"
  policy = "${file("${path.module}/templates/jenkins_role_policy.json")}"
}

resource "aws_iam_role_policy_attachment" "jenkins" {
  role       = "${aws_iam_role.jenkins.name}"
  policy_arn = "${aws_iam_policy.jenkins.arn}"
}
