# Vidlink

A small project to try out golang and buffalo. The system allows users to upload a video at /videos/new, runs a batch job 
in the background to HLS for a handful of resolutions, then allows users to share the unique url of the video. 

You can see the production app at https://coral-app-oyzm5.ondigitalocean.app/ unfortunately, it's hosted on a PaaS product 
by DigitalOcean that has less CPU horsepower than a gameboy color. If I have time, then I'll use mrsk to deploy it to 
a bare metal server with sufficient CPU to encode videos quickly. 

A demo video can be found here: https://coral-app-oyzm5.ondigitalocean.app/videos/be9047b8-65c5-49eb-b9c6-78509a2f53b7

There is almost no effort put into making this app safe. Use at your peril!!!


## Database Setup

It looks like you chose to set up your application using a database! Fantastic!

The first thing you need to do is open up the "database.yml" file and edit it to use the correct usernames, passwords, hosts, etc... that are appropriate for your environment.

You will also need to make sure that **you** start/install the database of your choice. Buffalo **won't** install and start it for you.

### Create Your Databases

Ok, so you've edited the "database.yml" file and started your database, now Buffalo can create the databases in that file for you:

```console
buffalo pop create -a
```

## Starting the Application

Buffalo ships with a command that will watch your application and automatically rebuild the Go binary and any assets for you. To do that run the "buffalo dev" command:

```console
buffalo dev
```

If you point your browser to [http://127.0.0.1:3000](http://127.0.0.1:3000) you should see a "Welcome to Buffalo!" page.

**Congratulations!** You now have your Buffalo application up and running.

## What Next?

We recommend you heading over to [http://gobuffalo.io](http://gobuffalo.io) and reviewing all of the great documentation there.

Good luck!

[Powered by Buffalo](http://gobuffalo.io)
