resource "aws_s3_bucket" "test_s3" {
  bucket = "test-mu-bucket"

  tags = {
    Name        = "test bucket"
    Environment = "Test"
  }
}
