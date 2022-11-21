resource "local_file" "test" {
  content  = "this is test"
  filename = "${path.module}/${var.local_file_path}"
}
