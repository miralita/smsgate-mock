Smsgate-mock

Very simple pure-go emulator for SMS gate (REST). Using local database as storage.

    $ docker build -t smsgate-mock .
    $ docker run --name smsgate-mock -v dbdata:/app/dbdata -d -p 8811:8811 smsgate-mock

See .env file for a configuration.
You can also overwrite .env file using environment variables.

Default port is 8811.

See Swagger for API documentation http://localhost:8811/swagger/index.html

Implemented methods:

* Add/edit/check/list/delete senders
* Add/list/search/check/delete messages
