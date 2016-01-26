Grant/authenticate access to json obj by temporary KPASS uuid. Keys can be reset quickly with RPASS uuid. To be implemented in restful api which formats keys for delivery to users and authenticates a super user for adding and deleting records.

Authority takes a string and stores it in a text file with the filename KPASS.RPASS, where each PASS is a UUID string. Read requests are authorized for queries with a matching KPASS. Filenames are refreshed with new UUIDs when the RPASS is presented.
