this is simple http server support download and upload implement by golang.
# functions:
- run with params -h port -d /tmp/upload to support http file server
- get / list all file in /tmp/upload
- get /test list all file in /tmp/upload/test
- get /test/test.txt to view file in chrome if the file is text file
- get /test/test.data to downlaod file if the file is not text file
- support upload file
-- 'curl -X PUT --upload-file @test.txt http://127.0.0.1:8000/test.txt' upload to test.txt
-- 'curl -X PUT --upload-file @test.txt http://127.0.0.1:8000/test1/test2.txt upload to  /test1/test2.txt