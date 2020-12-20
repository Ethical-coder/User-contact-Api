package main


import (
	"fmt"
	"log"
	"net/http"
	"time"
	"io/ioutil"
    "encoding/json"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    //uncomment code below for using ping 
    //"go.mongodb.org/mongo-driver/mongo/readpref"
    "context"

)
//global variables declaration to use it with different functions
var client *mongo.Client
//data to be feeded in the mongodb server
type people struct{
	name string
	DOB string
	number string
	email string
	id primitive.ObjectID 
	timestamp time.Time

}
 //data structure for recieving user creation req
type create_People struct{
	Name string
	DOB string
	Number string
	Email string
}
//contact data
type contact_data struct{
	Person1_id string
	Person2_id string
	timestamp string

}

type error_mess struct{
	message string
}

type people_data struct{
	name string
	DOB string
	number string
	email string
	id string
	timestamp time.Time

}





//Post reqeuest for creating users 
func create_user(w http.ResponseWriter, r *http.Request){
	//r.ParseForm()
	
	//converts the body of the post api to byte buffer
	switch r.Method{

	//POST method for creating users
	case "POST":
		var buffer,_=ioutil.ReadAll(r.Body)
		
		//creating data structure for storing api body
		var user_entered_data create_People
		
		//converting json bytes to data structure in golang 
		json.Unmarshal(buffer,&user_entered_data)

		fmt.Println("user name - ",user_entered_data.Name," , user dob - ",user_entered_data.DOB," ,user no. - ",user_entered_data.Number,"  user email = ",user_entered_data.Email)
		//person:=people{user_entered_data.Name, user_entered_data.DOB,user_entered_data.Number,user_entered_data.Email,last_id,time.Now()}
		//response,_:=json.Marshal(person)
		//storing data in database
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		//making instance of user database
		user_database := client.Database("contact_user_database").Collection("Users")
		res, err := user_database.InsertOne(ctx, bson.M{"name": user_entered_data.Name, "DOB": user_entered_data.DOB,"number":user_entered_data.Number,"email":user_entered_data.Email,"timestamp":time.Now().UTC()})

		if(err!=nil){
		er1:=error_mess{"Error in interacting with database"}
		json.NewEncoder(w).Encode(er1)		
		return	
		}
		json.NewEncoder(w).Encode(res)
		fmt.Println("user id ======",res)
		defer cancel()
	

	//GET method for gettign users using id
	case "GET":
		id:=r.URL.Path[len("/User/"):]
		docId, _ := primitive.ObjectIDFromHex(id)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		user_database := client.Database("contact_user_database").Collection("Users")
		cur, _ := user_database.Find(ctx, bson.D{})
		defer cur.Close(ctx)

		for cur.Next(ctx) {
		   var result bson.M
		   cur.Decode(&result)
		   // do something with result....
		   if(result["_id"]==docId){
		   		json.NewEncoder(w).Encode(result)		

		   }
		   
		}
		defer cancel()

	}
}

//update the contacts between 2 ids
func update_contact(w http.ResponseWriter, r *http.Request){
	switch r.Method{

	case "POST":
		r.ParseForm()
		
		//converts the body of the post api to byte buffer
		var buffer,_=ioutil.ReadAll(r.Body)
		
		//creating data structure for storing id of persons in contact
		var contact_data_id contact_data
		
		//converting json bytes to data structure in golang 
		json.Unmarshal(buffer,&contact_data_id)


		//object does not gets created but no error message displayed
		if(string(contact_data_id.Person1_id)==string(contact_data_id.Person2_id)){
			w.Write([]byte(`{"message":"Both the user IDs can't be same in the contact"}`))
			return	
		}

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		//making instance of user database
		collection := client.Database("contact_user_database").Collection("ids")

		//converts the string in YYYY-MM-DD format to timestamp object
		t,_:=time.Parse("2006-12-31",contact_data_id.timestamp)

		res,err:=collection.InsertOne(ctx, bson.M{"id1": contact_data_id.Person1_id, "id2":contact_data_id.Person2_id ,"timestamp":t})

		if(err!=nil){
		w.Write([]byte(`{"message":"Error in interacting with database"}`))
		return	
		}
		json.NewEncoder(w).Encode(res)

		

		
	case "GET":
		//declaring people array
		var people_data_sent []people

		r.ParseForm()
		//extracting form data set
		user_id:=r.Form["user id"][0]
		docId, _ := primitive.ObjectIDFromHex(user_id)
		time_of_infection,_:=time.Parse("2012-12-31",r.Form["infection_timestamp"][0])
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		
		//declaring database pointers
		infection_data := client.Database("contact_user_database").Collection("ids")
		user_data := client.Database("contact_user_database").Collection("Users")

		//creating pointer to id data base
		cur_infection_data,_:=infection_data.Find(ctx, bson.D{})
		cur_user_data,_:=user_data.Find(ctx, bson.D{})
		defer cur_infection_data.Close(ctx)
		//checking all possible conntacts for given id 
		for cur_infection_data.Next(ctx) {
		   var resultm bson.M
		   cur_infection_data.Decode(&resultm)
		   // 
		   end:=time_of_infection.AddDate(0,0,-14)
		   start:=time_of_infection
		   contact_date,_:=time.Parse("2006-12-31",resultm["timestamp"])
		   if (end.After(contact_date) && start.Before(contact_date)){
			   	if(resultm["Person1_id"]==docId){
			   		cur_user_data,_=user_data.Find(ctx, bson.D{})	
			   				for cur_user_data.Next(ctx) {
							   var result people
							   var result1 bson.M
							   cur_user_data.Decode(&result)
							   cur_user_data.Decode(&result1)
							   // do something with result....
							   if(result1["_id"]==docId){
							   		people_data_sent=append(people_data_sent,result)		

		   						}  
							}
		   	}

		   }
		   		   
		}
		json.NewEncoder(w).Encode(people_data_sent)

		fmt.Println(r.Form)
}
}




//main function that cotrols the working of all the functions
func main(){
	fmt.Println("> Application has started.")
	//creates a mongodb client
	fmt.Println("> Connecting to Database ...")
	client,_=mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/"))
	//context time out after how much time err should be popped
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//connecting to client
	client.Connect(ctx)
	fmt.Println("> Database connected.")
	//All the commented code below is used for checking conectivity with client
	/*err = client.Ping(ctx, readpref.Primary())
	if err!=nil{
		log.Fatal(err)
	}*/
	//databases, _:= client.ListDatabaseNames(ctx,bson.M{})

	

	//different routes
	fmt.Println("> Waiting for requests ...")
	http.HandleFunc("/User/",create_user)
	http.HandleFunc("/contacts",update_contact)

	log.Fatal(http.ListenAndServe(":8001",nil))

	defer cancel()

}

