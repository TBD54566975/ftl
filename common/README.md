This package and all subpackages should only contain code that is common to both the end user runtime and the FTL server
itself.

Adding additional dependencies will increase end user build times and binary sizes, so it is important to keep this package
as small as possible.