# Apple SG Inventory Notifier

Telegram bot that notifies you when the Apple Store in Singapore has the product you want in stock.

![Screenshot of Telegram bot](./assets/demo.png)

> [!IMPORTANT]
> Refer to `tracking_request.json` to view the models that are currently being tracked.
>
> Feel free to open an issue or submit a pull request to add more models to track.

## Tech Stack

|                                                           | **Purpose/Rationale**                                                             |
| --------------------------------------------------------- | --------------------------------------------------------------------------------- |
| [**Go**](https://golang.org/)                             | Easy setup, easy to use, my grug brain like ([*context*](https://grugbrain.dev/)) |
| [**Telegram**](https://core.telegram.org/bots/api)        | Go-to messaging app commonly used by Singaporeans, supports broadcast channels    |
| [**Cloud Function**](https://cloud.google.com/functions)  | Generous free tier, pay for what you use                                          |
| [**Cloud Scheduler**](https://cloud.google.com/scheduler) | Setup cron jobs with HTTP POST                                                    |
| [**MongoDB Atlas**](https://www.mongodb.com/cloud/atlas)  | Data persistence to easily track alerts sent, easy setup                          |

> [!NOTE]
> For the full details, feel free to read in the [devlog](https://github.com/yusufaine/apple-notifier/blob/main/assets/devlog.md).

### Local development and testing

```bash
FUNCTION_NAME=apple_notifier go run cmd/localcf/main.go
# Alternatively, use the .env file
# env $(cat .env | xargs) go run cmd/localcf/main.go
```

### Deploying

```bash
FUNCTION_NAME=apple_notifier go run cmd/zipcf/main.go
# Alternatively, use the .env file
# env $(cat .env | xargs) go run cmd/zipcf/main.go
```

1. Zip the cloud function using the command above
2. Upload the zip file to the cloud function
3. Set the cloud scheduler to send a HTTP POST request with the `tracking_request.json` as the body to the cloud function

#### Tracking other Apple Stores

You can also run your own instance of the bot to track models from other countries, though I've only managed to ensure that this works for Apple stores in Singapore and Hong Kong.

> Hong Kong Apple stores example

```json
{
  "abbrev_country": "hk",
  "country": "hong kong",
  "models": ["MU7J3ZP/A", "MU7E3ZP/A", "MU793ZP/A"]
}
```

## Credits

There was another Github repo I saw that did something similar but used a desktop application to notify the users that was built in Go with Fyne, I believe. I can't seem to find it anymore but I'd like to credit the author for the inspiration and providing the endpoint to query the Apple Store inventory.
