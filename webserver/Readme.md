A webserver in golang which has multiple pages
1. Genric Home page which displays producst and services along with log in and sign up option. Product lists will be fetched from an api without authentication
2. User login page it will be a kind of form which will  take user name and password as input and request to aut API to get authentication, if authenticated it will get succes and auth key. which will be used to get data in user home page else it will show wrog password window
3. User home page. it will be a tempalte whcih will open and details like uder name etc will be fetched from another api along with auth key. inreturn it will get user name and other details.
4. Sign up page which will be a form for user details and once entered with submit it will send post request to userADD api 
