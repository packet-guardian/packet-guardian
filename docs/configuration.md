# Configuration

The configuration file is written in [TOML](https://github.com/toml-lang/toml).
All available options are given in the sample configuration along with their
defaults and a short explanation.

**NOTE**: Except for the first letter, configuration setting names are CASE
SENSITIVE. Eg, `AllowManualRegistrations` is not the same as
`allowmanualregistrations`. However it is the same as
`allowManualRegistrations`. It is recommended to use the same case as the sample
configuration.

Here's a quick run down of the main sections of the configuration:

- **Core**: This is for system-wide settings that basically didn't fit anywhere
  else. In particular this contains the company name, website title, and domain
  name.
- **Logging**: Where and how much to log. The default logging level is "notice".
  Standard syslog levels are valid plus `fatal`.
- **Database**: Where and how to store data. Currently only SQLite is support
  but there are plans to include support for PostgreSQL and MySQL.
- **Registration**: How to handle device registrations and setting defaults such
  as how many devices each user can have and the method used to expire a device.
- **Leases**: Enable/disable lease history and settings that pertain to it.
- **Guest**: Guest specific registration settings. It has many of the same types
  of settings as Registration, but is only for "guest" users. Here is also where
  you specify the method used to verify guest users. Currently only Twilio is
  support but more are coming soon.
    - **Email**: Settings pertaining to email verification for guest
      registrations. Not implemented yet.
    - **Twilio**: Settings pertaining to using Twilio SMS messaging to verify
      users. One text message is sent with a verification code for the user to
      enter.
    - **SMSEagle**: Allows sending text messages via SMSEagle with a
      verification code.
- **Webserver**: How the webserver should behave. Ports, addresses, TLS
  settings, and session settings.
- **Auth**: How to authentication non-Guest users.
    - **LDAP**: Use LDAP/Active Directory authentication. Currently only AD
      authentication is supported.
    - **Radius**: Use a Radius server.
    - **CAS**: Use a central authentication server. Packet Guardian will not
      redirect users to a CAS login page. Instead it will acquire a login token
      and perform a login request on behalf of the user. The returned ticket is
      verified and then forgotten. This method does not allow/support single
      sign on.
