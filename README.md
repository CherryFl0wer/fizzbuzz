## FizzBuzz exercice 

### Goal

"Exercise: Write a simple fizz-buzz REST server.

"The original fizz-buzz consists in writing all numbers from 1 to 100, and just replacing all multiples of 3 by ""fizz"", all multiples of 5 by ""buzz"", and all multiples of 15 by ""fizzbuzz"".

The output would look like this: ""1,2,fizz,4,buzz,fizz,7,8,fizz,buzz,11,fizz,13,14,fizzbuzz,16,...""."

Your goal is to implement a web server that will expose a REST API endpoint that:

Accepts five parameters: three integers int1, int2 and limit, and two strings str1 and str2.
Returns a list of strings with numbers from 1 to limit, where: all multiples of int1 are replaced by str1, all multiples of int2 are replaced by str2, all multiples of int1 and int2 are replaced by str1str2.

The server needs to be:

Ready for production
Easy to maintain by other developers


Bonus: add a statistics endpoint allowing users to know what the most frequent request has been. This endpoint should:

Accept no parameter
Return the parameters corresponding to the most used request, as well as the number of hits for this request"

### How to run 

Run can be made by running `make start` which will build the image of the app and run both the cache and monitoring instances.

### Run tests 

`make tests` is enough 

### Done 
- [X] Request a fizzbuzz array depending on parameters (Form or JSON) `POST /fizzbuzz`
- [X] Return top requested fizzbuzz request on a `GET /metrics` 
- [X] Metrics for the app exported to prom
- [X] Simple swagger 
- [X] Partials Tests

### Missing
- [ ] Dashboard with alert
- [ ] Integration Tests for api
- [ ] Swagger complete integration with tests
- [ ] Healthcheck for app 
- [ ] No tests for metric service and metrics controller