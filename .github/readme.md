# Disposable Services Registry

A register of transient disposable services.

## Features

- Lists of valid disposable domains
- Lists of valid throwaway phone numbers

#### Get all items

| Parameter | Description                |
| :-------- | :------------------------- |
| `https://raw.githubusercontent.com/complexorganizations/disposable-services-registry/main/assets/disposable-domains` | **Domains** |
| `https://raw.githubusercontent.com/complexorganizations/disposable-services-registry/main/assets/disposable-telephone-numbers` | **Telephone numbers** |


## FAQ

#### How frequently are the listings updated?

The lists are updated on a daily basis using github actions.

#### Is it possible for me to upload a list that I discovered?

To add a new list, please send a PR; in case you're wondering, all lists must be public in order to be scraped.

## Manually update the lists

Duplicate the project.

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

[Apache License Version 2.0](https://raw.githubusercontent.com/complexorganizations/disposable-services-registry/main/.github/license)


## Used By

This project is used by the following companies:

-
-
