# db2xlsx

Database structure export to XLSX. This software connects to a MySQL (or
compatible, e.g. MariaDB) database and extracts a list of all the tables
and their columns and exports them to an XLSX file:

* One XLSX sheet per database table
* One XLSX row per database column

It is intended for a very simple use case where a database needs to be
visualised in a user-friendly format (XLSX is more friendly to a lot of
users, especially those who are not DBAs).
