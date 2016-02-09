#Notes on Shipyard Work

> You must be logged in as the default admin user to execute these steps.
> `oc login -u admin -p test`

###Create `shipyard` Project


```sh
oc new-project shipyard
```

###Create Service Account Named `foreman` under `shipyard` Project

Service Accounts are API objects. Can be created with `POST` requests or by using the `oc` cli.

Example Service Account object file. Copy this code into `sa.json`

```json
{
  "apiVersion": "v1",  
  "kind": "ServiceAccount",  
  "metadata": {
    "name": "foreman"
  }
}
```

To create this service account 

```sh
oc create -f sa.json
```

###Grant `cluster-admin` Role to `foreman` User

> If `foreman` needs `cluster-admin` role on all projects then

You must ssh into the vagrant box with `vagrant ssh` and use the `oadm` cli for this.

```sh
oadm policy add-cluster-role-to-user cluster-admin foreman
```

>If `foreman` only needs `cluster-admin` Role on `shipyard` then

```sh
oc policy add-role-to-user cluster-admin foreman -n shipyard
```

###Make `foreman` an admin on `shipyard` project

```sh
oc policy add-role-to-user admin foreman -n shipyard
```

###Make a Binary Build

```sh
oc new-build --binary --name=origin-management-poc
```

###Start the Build from `config` Directory

```sh
oc start-build origin-management-poc --from-dir=config --follow
```

###Deploy the Build Using `oc new-app`

```
oc new-app origin-management-poc
```