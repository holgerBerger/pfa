@0xbb46324ce49ee65e;
using Go = import "/go.capnp";
$Go.package("pfalib");
$Go.import("zombiezen.com/go/capnproto2/example");

struct FileEntry {
  name  @0  :Text;
  uid   @1  :Int64;
  gid   @2  :Int64;
  owner @3  :Text;
  group @4  :Text;
  mtime @5  :Int64;
  ctime @6  :Int64;
  atime @7  :Int64;
  mode  @8  :Int64;
  size  @9  :Int64;
  id    @10 :Int64;
}

struct FileSegment {
  id    @0  :Int64;
  size  @1  :Int64;
}

struct DirectoryEntry {
  name  @0  :Text;
  uid   @1  :Int64;
  gid   @2  :Int64;
  owner @3  :Text;
  goup @4  :Text;
  mtime @5  :Int64;
  ctime @6  :Int64;
  atime @7  :Int64;
  mode  @8  :Int64;
}
