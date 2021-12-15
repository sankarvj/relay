# relay-worker

This worker starts "RunReminderListener" which polls the redis list for the latest reminder to be processed. The reminders will get added from the core app, the listener simply checks the latest work to be execute and process if the time elapsed.

## Points to note:

- What happens if more number of jobs clogs at some point of time.
    The channels will process the jobs one by one and call the core API/Method for further processing. Horizontal scaling will help the core process and redis poll. 


