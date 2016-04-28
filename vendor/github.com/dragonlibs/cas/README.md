Dragonlibs CAS
==============

This library performs simple authentication against a CAS server without having to redirect the user. The library will scrape a login token from the server's login page and perform a request on behalf of the user. The client application will receive an object that contains the attributes of the user as received by the server.

Note: This library does NOT support single sign on or single log off. It's an isolated request. The client application is responsible for managing logged in and logged out status.

This library was written to accommodate a need in one of my applications namely I wanted a system of successive authentication methods trying each one in turn. Doing CAS the normal way would make this difficult as one of two solutions would need to be used. 1) Have a separate login for non-cas logins. This however would be confusing. 2) Redirect to CAS after a failed login where the user would have to enter their credentials again. This however is likewise confusing and user unfriendly.

I make no guarantees that this library will work for you. However, if an issue is found please let me know.

This library assumes the CAS login page has an HTML input field with the name "lt" as per the specification. If it does not, this library will not work.
