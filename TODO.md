# Various TODO and Crib Notes

In no particular order ...

## TRIBE Integration
Get it hooked up to TRIBE for consumption-based billing and RBAC.

Bonus points for demonstrating TRIBE being hooked up to stripe or gocardless.com

## Grafana Dashboards
Nice eye-candy

## Other services
Demonstrate other services with different providers to choose from:

    - transcribe

    - translate

## Distributed Attenuation
A redis implementation of the attenuator so that multiple of these can be co-ordinated

## Kubernetes + HELM
Raw kubernetes yaml

Helm charts

## S3 (and other storage) integration for saving requests and responses

## Admin Console
Using the config API to change behaviour in real time

## Non-HTTP Things That Can Fail

    - DNS lookups (see https://isitdns.com)

    - network timeout

    - connection refused

## Notes
This is a labour of love, something I've wanted to build for a long time, and I'm building it anyway because it will really really help me with the various limitations I encounter when trying to scrape content and exceed the limits they put in place to stop me (e.g. companies house, or the classic cars site I did for Charlie)

I'm starting to think that there are (at least) seven separate use cases for commercialising this:

    - enapsulating all retry / failure logic behind a single interface

    - testing / validating fault handling in code

    - SRE training / playbook creation (i.e. known faults produce known graphs with known resolution)

    - External monitoring
      Poke a service from external and STB if it matches a given failure profile

    - Saas rate-limiting (broker mode)
      (you have a Saas, you want to perform rate-limiting)

    - Saas billing (broker mode)
      (you have a Saas, you want to charge your customers on some basis related to consumption)

    - RBAC (broker mode)
      (You have a REST service, you want to restrict people from doing certain things according to their role / permissions)

Of these, I think that SRE training is the most interesting (because it is a VERY hard role to train people in - you can't teach experience and lore).  Perhaps that is the £££opportunity here, because it's such a hard thing to do:

    - create systems with known fault modes

    - hook up the graphs / logs etc

    - teach people how to interpret the graphs and consult the playbooks

