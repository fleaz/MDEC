# mdec
Multi Domain Email Confugration


## Problem and requirements
Adding a new mail account to your mail client is always way more pleasant when the client autodetects all the settings
for incoming and outgoing server, the port, which SSL method to use and so on. All big providers like e.g. GMail have
this, but most of the time when you run your own mailserver this gets left behind because it's probably not as much of a
deal when you just run your own mailserver for youself and run a hand full of devices that never change. But when you
host mail for friends and family or are like me and run a small hosting business where have have mails hosted for
customers, you spend a lot of your time telling people how to configure a mailclient. This get's especially challenging
when they use a client you have never touched yourself and don't know where all the settings are.  Therefore I wanted a
relatively simple solution that is capable of handling multiple domains. I know there exist some scripts/tools but they
are either only good for a single domain or have to many features that I don't need and are therefore harder to setup
(AutoMX looks like good but they need a database which I don't want to maintain just for email autodiscovery)

## mdec setup
Todo..

## DNS
### Single domain setup
If you want to use mdec for a single domain, let's say `example.com`, you need to add two DNS records (A,AAAA or CNAME
doesn't matter) which point to a webserver:
  * autoconfig.example.com
  * autodiscover.example.com
Then you need to get a SSL certificate for `autodiscover.example.com`. Now scroll down to *Webserver setup* and
configure the webserver to serve the XML files.

### Multi domain setup
Here comes the actual fun part. Lets say you host emails both for `example.com` and `awesome.org`. Like in the example
above you need one domain, lets call it the main domain, which has the actual webserver runnig which serves the mdec
tool. We will use example.com as our main domain for this example. Configure these two DNS records:
  * autoconfig.example.com
  * autodiscover.example.com

Now for `awesome.org` (and every other domain) we also need to add two records, but the nice part is, that no matter how
many domains we host, we neither have to touch the dns of our main domain, touch the webserver config or get any
additional SSL certificates. Neat!

You just need these two records:
  * CNAME for `autoconfig.awesome.org` which points to `autoconfig.example.com`
  * SRV record for `_autodiscover._tcp.awesome.org` with the value `10 0 443 autodiscover.example.com`

We can use a simple CNAME for autoconfig because it's plain HTTP so the webserver will serve our content no matter with
which domain we make a request. The problem is that autodiscover on the other hand requires HTTPS, so if we would also
just use a CNAME for the autodiscover domain, we ould need to have a valid SSL certificate for every domain we host on
the webserver which serves mdec. Nowadays with fancy webserves like traefik and caddy which can auto-issue new Let's
Encrypt certificate for new domains, this theoretically isn't much of a problem anymore but I still wanted to keep
things simple and have a bunch of certificates just for a single autodiscover requests of a new email client seems
bloated so I went with the SRV record method. This way we can tell the client which requested the autodiscovery service
that it's actually hosted on `autodiscover.example.com`. This way we get proper redirect to our main domain and only
need a single certificate and the webserver only needs to listen on the main domain.

## Webserver setup
Todo...


## Configuration
mdec is managed via the *config.yaml* file. There are some global settings to define the listening address of the tool
and the log level as well as a entry called *domains*. This key contains a key/value pair of domain names, namely the
domains you host on your mailserver. Because when hosting multiple domains on the same mailerver most of the settings
will be the same, therefore there is a special entry with the key *default*. This entry defines the fallback values for
all domains that are not explicitly defined in the file and also gets merged with all explicitly defined domain entries,
so if a domain only has one different value, you just need to define it and all the other values gets inherited from the
*default* entry.

## Supported protocols

Like everywhere else in IT we have [many competing standards](https://xkcd.com/927/). I tried to cover all of the
important ones so this tools works for all major desktop and mobile clients. All information below is to my best
knowledge and if something is wrong I am happy for issues of pull requests!

### Autoconfig
This is the standard developed by Mozilla and is used e.g. in Thunderbird.  Autoconfig tries to contact
`http://autoconfig.<your-emaildomain>` and requests an XML file from `/mail/config-v1.1.xml` while providing the
emailaddres they requested via the `emailaddress` GET-Parameter. They request only plain HTTP and don't accept a
redirect to https so you need to deliver it plain.

Documentation for the file format can be found in the
[Mozilla Wiki](https://wiki.mozilla.org/Thunderbird:Autoconfiguration:ConfigFileFormat)

### Autodiscover 
This was the old standard invented by Microsoft in e.g. Outlook. From what I know this is no longer used since 2016.
There this is not implemented here.

### Autodiscover V2
This is the successor to the old Autodiscover protocol. But apparantlt only supports Office365 and no longer plain IMAP
or POP3. Therefore this can't be used for selfhosted mail servers. Thanks Microsoft!

### Mobileconfig
Todo...
