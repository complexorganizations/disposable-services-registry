# Disposable Services Registry

A registry of temporary disposable services.

***Disposable Services Registry is not yet complete. You should not rely on this code. It has not undergone proper degrees of security auditing and the protocol is still subject to change. We're working toward a stable 1.0.0 release, but that time has not yet come. There are experimental snapshots tagged with "0.0.0.MM-DD-YYYY", but these should not be considered real releases and they may contain security vulnerabilities (which would not be eligible for CVEs, since this is pre-release snapshot software). If you are packaging Disposable Services Registry, you must keep up to date with the snapshots.***

## Features

- Validate domain lists, containg all valid disposable domains
- valid phone numbers lists, all valid ones


#### Get all items

| Parameter | Description                |
| :-------- | :------------------------- |
| `https://raw.githubusercontent.com/complexorganizations/disposable-services-registry/main/assets/disposable-domains` | **Domains** |
| `https://raw.githubusercontent.com/complexorganizations/disposable-services-registry/main/assets/disposable-telephone-numbers` | **Telephone numbers** |


## FAQ

#### How frequently are the listings updated?

Using github actions, the lists are updated on a daily basis.

#### Is it possible for me to upload a list that I discovered?

Please submit a PR to add a new list; if you're wondering, all lists must be public in order to be scraped.

## Update the lists

Clone the project

```bash
  git clone https://github.com/complexorganizations/disposable-services-registry
```

Go to the project directory

```bash
  cd disposable-services-registry
```

Install the required dependencies and compile the code

```bash
  go build .
```

Refresh the lists and make a new one

```bash
  ./disposable-services-registry -update
```


## Roadmap

- Make it better by adding additional listings.


## Feedback

Please utilize the github repo conversations to offer feedback.


## Support

Please utilize the github repo issue and wiki for help.


## Authors

- [@prajwal-koirala](https://github.com/prajwal-koirala)

Open Source Community


## License

[MIT](https://raw.githubusercontent.com/complexorganizations/disposable-services-registry/main/.github/license)


## Used By

This project is used by the following companies:

- 
- 
