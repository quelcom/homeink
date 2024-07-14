# Homeink

Transforming my old Kindle DX into a smart(ish) e-ink clock.

## Intro

Back in 2010 I bought an Amazon Kindle DX, thinking that the large screen would be ideal for reading technical books in PDF format. Even though reading such books was *doable* overall, it required few compromises. In some cases, PDF books required a bit of preprocessing in order to crop empty margins. Despite that, PDFs with condensed charts or images were still difficult to interpret, and the zooming and navigation capabilities of the device were a bit underwhelming, especially with heavy files. In the end, I used the Kindle DX for reading mostly sci-fi books, and that was perfect especially when I learned how to hack the device and install [KOReader](https://github.com/koreader/koreader). That lasted for years until the battery started to give up due to its age. In 2021 I bought a Kobo Libra 2 and the Kindle went right into a drawer. I refused the idea of consiering the Kindle e-waste and that I should just get rid of it, so it went to a drawer waiting for a new opportunity.

And the opportunity came: I was looking to buy a large (around 10" or 12") displays in order to display a Home Assistant dashboard, while I remembered I still had my old Kindle. The major blocker is that the Kindle DX does not have any ethernet or wi-fi connecitivy (only networking available is the built-in 3G radio for Amazon Whispernet, which has been defunct for years). On the other hand, a Raspberry Pi Zero 2 W has all the networking I need and is tiny. Considering I already hacked the Kindle back in the days, I started to explore the idea of coupling them and see if I could control the Kindle display from the Pi Zero.

## Preparations

### Kindle

I installed the latest version of the USBNetwork hack from [this post in Mobileread](https://www.mobileread.com/forums/showthread.php?t=88004). Installation is straight-forward, just following the instractions and making sure I am getting the correct .bin file for the DX model.

I also installed [FBInk](https://github.com/NiLuJe/FBInk) to control the e-ink display, which is a great alternative to the default `eips` command.

### Pi Zero

I installed DietPi because I had an unused 2 GB card unused and I didn't want to buy a new one. I roll with the defaults, but I added a new network interface in order to being able to SSH to the Kindle over [RNDIS](https://en.wikipedia.org/wiki/RNDIS):

```
dietpi@DietPi:~$ cat /etc/network/interfaces.d/usb0
auto usb0
iface usb0 inet static
address 192.168.2.1
netmask 255.255.255.0
```

I also adjusted the cpu governor to powersave and few tweaks here and there, but nothing important other than adding the new network interface.

## Software

### Homeink

The main software is called `Homeink` and is a Go application that gets installed in the Pi Zero device. In a nutshell:

- It provides a small wrapper to the FBInk tool, and it runs `fbink` commands over SSH thanks to the [goph](https://github.com/melbahja/goph) library.
- It is also a small HTTP server that currently exposes two endpoints available in the local network. At the moment there are two endpoints:
  - `/api/v1/screenshot` that accepts an image to be rendered in the display. Used to render weather forecasts from [Supersää](https://supersaa.fi).
  - `/api/v1/water`, that is used to show the daily water consumption in our home.

The HTTP server also exposes a Swagger UI under `/swagger/index.html`, which is useful to test the endpoints, especially the screenshot endpoint because it is a multipart request.
