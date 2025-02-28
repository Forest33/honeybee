<p align="center">
<img src="assets/honeybee.png" style="width:150px" width="150" alt="Honeybee - IoT Device Management System" />
<h1 align="center">Honeybee</h1>
<h4 align="center">IoT Device Management System</h4>
</p>

Honeybee is a powerful platform for automating and managing smart devices through Internet of Things (IoT) protocols. 
The application allows you to create flexible scripts to interact with various devices such as sensors, actuators, 
and other IoT gadgets. Using the built-in programming language Lua, you can implement any control logic by creating 
unique automation algorithms.

## Key Features

### Integration with Zigbee2MQTT

The app is perfectly suited for working with the zigbee2mqtt system. Through this integration, Honeybee provides complete 
control over all connected devices, allowing easy management from within a single application. However, Honeybee also 
supports operation independently of zigbee2mqtt, making it a versatile tool for any IoT scenario.

### MQTT Interaction

Honeybee actively communicates with MQTT brokers, enabling two-way communication between devices and the application. 
This means that you can receive event data from devices, process it according to your scripts, and send commands back into 
MQTT topics. For example, a script could monitor temperature changes from a sensor and automatically turn on an air 
conditioner when a certain threshold is reached.

### Event Notifications

The platform offers convenient tools for notifying users about important events. By integrating with the service [ntfy.sh](https://ntfy.sh), 
Honeybee can send push notifications directly to your smartphone or messages via Telegram. This ensures you're always 
aware of the status of your devices and overall system.

### Timers and Alarms

For automating processes at specific times or intervals, Honeybee offers timer, ticker, and alarm functions. 
These elements allow you to create scheduled scripts, significantly expanding the possibilities for configuring the 
behavior of your IoT infrastructure.

### JSON Serialization

Thanks to support for JSON serialization, you have the ability to easily exchange data between different components 
of the system. This simplifies handling complex structured data and facilitates interaction with external services and APIs.

## Install

```
# Run application by docker-compose
make
```