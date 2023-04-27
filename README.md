# http-attenuator

Http-attenuator runs as a API gateway providing:

    - attenuation

    - circuit-breakers / retry

    - middleware

    - request / response rewriting

    - consumption-based billing

    - RBAC and PBAC

    - metrics and monitoring

    - recording / saving requests and responses

(@JamesThatcher: some of this might remind you of 'parrot')

## MVP
    - [tick] forward proxy mode
      Needs integration with saving req/resp (a la BURP)

    - [tick] service broker mode (which also gives reverse proxy)

    - [tick] promana integration

    - [tick] basic "best" selection (for some definition of "best") by the service broker
      Currently supports `random` but you can pass it a `preferred` backend.
      Added `weighted`, which is good enough for a MVP.  But the whole backend selection
      could be a PhD project all by itself

    - [tick] recording / saving requests and responses
      TODO(john): Forward proxy needs this behaviour integrated
      TODO(john): drop gin, use net/http so we don't need different ports

    - circuit-break + attenuation + retry
      This is a solved problem, which I will integrate later (possibly wait until I've got
      the redis implementation of the clock)

    - [tick] grafana eye candy

    - [tick] demonstrate service broker

    - badly behaving server

    - useful service broker (storage: OneDrive, Drive, S3 - CRUD of same)

    - adaptive to 429
    
    - billing / tribe / cost in grafana (measure consumption)

    - kubernetes + HELM

## Why You Would Use This

### Gateway (forward proxy) Mode
Simply, it means that there is a whole bunch of stuff that you now do not have to do.

    - you don't need to build any error handling or retries into your client code

    - you don't need your client code to be directly connected to the internet

    - you don't need to build any rate-limiting plumbing - the attenuator gives you all the rate limiting for free

### Service Broker Mode
All of your requests - regardless of provider - return exactly the same JSON.

Your client code is much much easier, because it's all consistent.

Inb4 "lock in".  The very point of this is "lock in", with a FOSS solution.

Sometimes you might not know (or care) exactly which endpoint is the one you want to use,
or you want the _best_ endpoint (for some definition of _best_) to be chosen.

Imagine the case of video transcription and translation (which is the original reason why the
attenuator was built).

You have several providers for the transcription and translation:

    - aws
    - azure
    - google
    - whisper

and, depending on the parameters of the video (source language, target language, length) some of these are better than others.

Equally, it might be cheaper to send videos to AWS so you'd want to prioritise that and only fall back to the more expensive providers if AWS is down or slow for some reason.

You might want to route particular customers to particular services (e.g. because they may have services you want to consume that live in a particular cloud provider).

You might want to normalise all requests and responses, rather than having to write all of the client code that deals with the differences (e.g. some providers use millis for timestamps, others use seconds and others use ISO timestamps).

This is why broker mode exists.  Instead of specifying the exact endpoint like you would do in proxy mode, you just say which service you want to hit and the attenuator will route it according to the rules you've configured:

    http://{GATEWAY_ADDRESS}/api/v1/broker/{SERVICE}


### Failure Simulation Mode
In the event that you do want to build error handling and retries etc into your own code, the
attenuator can simulate the failure of endpoints in ways that you specify:

    - requests timing out

    - connection reset

    - garbled responses

    - different HTTP codes

and it can do this on a fixed (i.e. the problem always happens) or statistical (the problem
only happens n% of the time) basis.

This is particularly useful for being able to test your error-handling code in a development
environment, and allows you to automate end-to-end testing which proves the error handling
mechanisms you've built.

### Consumption-based Billing
Perhaps you're a service provider who wants to do consumption-based billing for your users.  For example, you might want to charge your transcription services per minute of video.  Or you might want to charge per-request.

You simply associate the rules with each service and you can measure consumption on any level you want.

### Monitoring and Logging
The attenuator integrates with Prometheus (and grafana) right out of the box, and comes complete with Grafana dashboards which are freely available in the Grafana marketplace.

It will integrate with ELK, LogStash, DataDog, StackDriver or any other logging provider - as well as integrating with SIEM tools like Splunk.

Point your prometheus config at http://{GATEWAY_ADDRESS}/metrics to start scraping.

If you add a header to your request of:

    X-Migaloo-Tag: foo

then this tag is available as a prometheus label.

### RBAC, Access Control and PBAC
Role-based access control is built-in to the attenuator so you can easily control access to services.

Furthermore, the RBAC built into the attenuator allows different service levels for different users, groups or organisations (e.g. org1 is allowed to do a maximum of 100 mins of video per day, but org2 can do a maximum of 300 mins per day).

In other words, you can build policies around access to services.

### Record and Save
When debugging HTTP-based problems, the ability to record traffic (requests and responses) including all headers etc is very useful.  This can be enabled via a simple config parameter

### TODO(john) Config API
Allow (some?) config to be set via API (e.g. 'enable record mode' or 'set debug log level') - similar to MBeans.


### TODO(john) Revese Proxy Mode
If you are a SaaS business, you can use reverse proxy mode to do consumption-based billing, without having to build all of that logic into your application.

## Getting started (forward proxy mode)

    docker build -t attenuator:latest .

    docker run -p 8888:8888 attenuator:latest

now just make your requests to

    http://localhost:8888/api/v1/gateway/{URL_AND_PARAMS}

and your requests will be proxed (with default retry and timeout settings).

For example, try fetching the google home page:

    curl http://localhost:8888/api/v1/gateway/https://google.com

## Simple Usage (forward proxy mode)

Just make your request to

    http://{GATEWAY_ADDRESS}/api/v1/gateway/{URL_AND_PARAMS}

The gateway will act as a forward proxy and will communicate with the service at `{URL_AND_PARAMS}`,
returning the response.

Error handling and retries are set according to the values in the `config.yml` file, either
globally or on a domain-by-domain basis.