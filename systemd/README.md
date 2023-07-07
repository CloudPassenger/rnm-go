## Service

### Install

#### 1. copy binary file

```bash
cp rnm-go /usr/bin/
```

#### 2. add service file

```bash
# copy service file to systemd
cp systemd/rnm-go.service /etc/systemd/system/
```

#### 3. add config file: config.json

```bash
# copy config file to /etc/rnm-go/
mkdir /etc/rnm-go/
cp config.example.json /etc/rnm-go/config.json
```

#### 4. enable and start service: rnm-go

```bash
# enable and start service
systemctl enable --now rnm-go
```

See [rnm-go.service](rnm-go.service)

### Auto-Reload
#### 1. enable and start
```bash
systemctl enable --now rnm-go-reload.timer
```

#### 2. customize
Execute the following command:
```bash
systemctl edit rnm-go-reload.timer
```

Fill in your customized values, for example:
```
# empty value means to remove the preset value
[Unit]
Description=
Description=Eight-Hourly Reload rnm-go service 

[Timer]
OnActiveSec=
OnActiveSec=8h
OnUnitActiveSec=
OnUnitActiveSec=8h
```

optionally do a daemon-reload afterwards:
```bash
systemctl daemon-reload
```
