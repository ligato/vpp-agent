# SQL like datastore

The package defines API for access data store using SQL. `Broker` interface allows to read and manipulate data.
`Watcher` provides functions for monitoring of changes in a datastore. 


Features:
-	User of the API has control about the SQL statements, types & binding which are passed to the `Broker`.  
-   Expressions:
    -  There ale helper functions that tends to avoid writing SQL strings. 
    -  It is up to the User if there will be used only expressions writen using helper function
    -  The user can even write by hand portions of SQL statements (sql.Exp helper function) 
       and combine them with other expressions. 
-	Optionally user can use reflection to simplify repetitive work with Iterators & GO structures
-   The is supposed to be reused among different databases. For each there is specific implementation.
