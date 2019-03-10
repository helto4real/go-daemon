# Setup the go-daemon applications in docker
This is how you setup the go-daemon in docker to get the following add-on apps
- Better presence
- (more to follow)

# Instructions
- Copy the all files under docker folder
- Check the volume under docker-compose.yaml file. Change the /opt/... etc to any path you desire the configuration file will apear
- do `docker-compose build` then `docker-compose up`
- exit 
- Open the configuration file at your host path. Should look like below. Add your own ip, if you run ssl set it to `true`. Most important you need a presistent token from homeassistant. Add it to `token`

```yaml
home_assistant:
  ip: 'your_ip_here:8123'           # Ip of your hass
  ssl: false                        # Set to true if hass using ssl
  token: 'homeasstant_token_here'   # Insert a long lived token here
```
## Better presence for people
If you are using better presence config the persons different devices. Need atleast one gps device_tracker and one or preferable one of each of wifi/bluetooth trackers.
```yaml

settings:
  tracking:
    just_arrived_time: 300                  # Default time from just arrived to home
    just_left_time: 60                      # Default time from just left to away
    home_state: "Home"                      # Default value for Home state
    just_left_state: "Just left"            # Default value for Just left state
    just_arrived_state: "Just arrived"      # Default value for Just arrived state
    away_state: "Away"                      # Default value for Away state

people:
  thomas:                                   #Each person has an id
    friendly_name: Thomas
    devices:
      - "device_tracker.thomas_phone_bt"    # The bluetooth tracker
      - "device_tracker.thomas_phone_gps"   # The gps tracker
      - "device_tracker.thomas_phone_wifi"  # The wifi (router) tracker
  jean:
    friendly_name: Jean
    devices:
      - "device_tracker.jeans_phone_bt"
      - "device_tracker.jeans_phone_gps"
      - "device_tracker.jeans_phone_wifi"
      
```

When all is configured correctly, do `docker-compose up`