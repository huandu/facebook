LÉAME.md
Un SDK de API Graph de Facebook en Golang
Estado de construcción GoDoc

Este es un paquete de Go que es totalmente compatible con la API Graph de Facebook con carga de archivos, solicitud por lotes y API de marketing. Se puede utilizar en Google App Engine.

La documentación de la API se puede encontrar en godoc .

Siéntase libre de crear un problema o enviarme una solicitud de extracción si tiene alguna pregunta, error o sugerencia de "cómo hacerlo" al usar este paquete. Haré todo lo posible para responder.

Instalar
Si go modestá habilitado, instale este paquete con go get github.com/huandu/facebook/v2. De lo contrario, llame go get -u github.com/huandu/facebookpara obtener la última versión de la rama principal.

Tenga en cuenta que, desde go1.14, las versiones incompatibles se omiten a menos que se especifique explícitamente. Por lo tanto, se recomienda actualizar la ruta de importación github.com/huandu/facebook/v2cuando sea posible para evitar cualquier posible error de dependencia.

Uso
Inicio rápido
Aquí hay una muestra que lee mi nombre de Facebook por uid.

package main

import (
    "fmt"
    fb "github.com/huandu/facebook/v2"
)

func main() {
    res, _ := fb.Get("/538744468", fb.Params{
        "fields": "first_name",
        "access_token": "a-valid-access-token",
    })
    fmt.Println("Here is my Facebook first name:", res["first_name"])
}
El tipo de reses fb.Result(también conocido como map[string]interface{}). Este tipo tiene varios métodos útiles para decodificar resa cualquier tipo de Go de forma segura.

// Decode "first_name" to a Go string.
var first_name string
res.DecodeField("first_name", &first_name)
fmt.Println("Here's an alternative way to get first_name:", first_name)

// It's also possible to decode the whole result into a predefined struct.
type User struct {
    FirstName string
}

var user User
res.Decode(&user)
fmt.Println("print first_name in struct:", user.FirstName)
Si un tipo implementa la json.Unmarshalerinterfaz, Decodeo DecodeFieldla usará para descomponer JSON.

res := Result{
    "create_time": "2006-01-02T15:16:17Z",
}

// Type `*time.Time` implements `json.Unmarshaler`.
// res.DecodeField will use the interface to unmarshal data.
var tm time.Time
res.DecodeField("create_time", &tm)
Leer un userobjeto gráfico con un token de acceso válido
res, err := fb.Get("/me/feed", fb.Params{
     "access_token": "a-valid-access-token",
})

if err != nil {
    // err can be a Facebook API error.
    // if so, the Error struct contains error details.
    if e, ok := err.(*Error); ok {
        fmt.Printf("facebook error. [message:%v] [type:%v] [code:%v] [subcode:%v] [trace:%v]",
            e.Message, e.Type, e.Code, e.ErrorSubcode, e.TraceID)
        return
    }

    // err can be an unmarshal error when Facebook API returns a message which is not JSON.
    if e, ok := err.(*UnmarshalError); ok {
        fmt.Printf("facebook error. [message:%v] [err:%v] [payload:%v]",
            e.Message, e.Err, string(e.Payload))
        return
    }

    return
}

// read my last feed story.
fmt.Println("My latest feed story is:", res.Get("data.0.story"))
Leer un gráfico searchpor página y decodificar fragmentos de mapas
res, _ := fb.Get("/pages/search", fb.Params{
        "access_token": "a-valid-access-token",
        "q":            "nightlife,singapore",
    })

var items []fb.Result

err := res.DecodeField("data", &items)

if err != nil {
    fmt.Printf("An error has happened %v", err)
    return
}

for _, item := range items {
    fmt.Println(item["id"])
}
Uso AppySession
Se recomienda usar Appy Sessionen una aplicación de producción. Proporcionan más control sobre todas las llamadas a la API. También pueden hacer que el código sea más claro y conciso.

// Create a global App var to hold app id and secret.
var globalApp = fb.New("your-app-id", "your-app-secret")

// Facebook asks for a valid redirect URI when parsing the signed request.
// It's a newly enforced policy starting as of late 2013.
globalApp.RedirectUri = "http://your.site/canvas/url/"

// Here comes a client with a Facebook signed request string in the query string.
// This will return a new session from a signed request.
session, _ := globalApp.SessionFromSignedRequest(signedRequest)

// If there is another way to get decoded access token,
// this will return a session created directly from the token.
session := globalApp.Session(token)

// This validates the access token by ensuring that the current user ID is properly returned. err is nil if the token is valid.
err := session.Validate()

// Use the new session to send an API request with the access token.
res, _ := session.Get("/me/feed", nil)
De forma predeterminada, todas las solicitudes se envían a los servidores de Facebook. Si desea anular la URL base de la API con fines de prueba unitaria, simplemente configure el Sessioncampo correspondiente.

testSrv := httptest.NewServer(someMux)
session.BaseURL = testSrv.URL + "/"
Facebook devuelve la mayoría de las marcas de tiempo en un formato ISO9601 que Go's no puede analizar de forma nativa encoding/json. La configuración RFC3339Timestamps trueen Sessiono en el nivel global hará que Facebook solicite las marcas de tiempo RFC3339 adecuadas. RFC3339 es lo que encoding/jsonse espera de forma nativa.

fb.RFC3339Timestamps = true
session.RFC3339Timestamps = true
Establecer cualquiera de estos en verdadero hará date_format=Y-m-d\TH:i:sPque se envíe como un parámetro en cada solicitud. La cadena de formato es una representación PHP date()de RFC3339. Hay más información disponible en este número .

Usar pagingcampo en respuesta
Algunas respuestas de Graph API usan una estructura JSON especial para proporcionar información de paginación. Úselo Result.Paging()para recorrer todos los datos en tales resultados.

res, _ := session.Get("/me/home", nil)

// create a paging structure.
paging, _ := res.Paging(session)

var allResults []Result

// append first page of results to slice of Result
allResults = append(allResults, paging.Data()...)

for {
  // get next page.
  noMore, err := paging.Next()
  if err != nil {
    panic(err)
  }
  if noMore {
    // No more results available
    break
  }
  // append current page of results to slice of Result
  allResults = append(allResults, paging.Data()...)
}
Leer la respuesta de Graph API y decodificar el resultado en una estructura
La API Graph de Facebook siempre usa claves de mayúsculas y minúsculas en la respuesta de la API. Este paquete puede convertir automáticamente nombres de campo de estructura de estilo de caja de serpiente a estilo de caja de camello de Go.

Por ejemplo, para decodificar la siguiente respuesta JSON...

{
  "foo_bar": "player"
}
Uno puede usar la siguiente estructura.

type Data struct {
    FooBar string  // "FooBar" maps to "foo_bar" in JSON automatically in this case.
}
La decodificación de cada campo de estructura se puede personalizar mediante la cadena de formato almacenada bajo la facebookclave o la clave "json" en la etiqueta del campo de estructura. Se recomienda la facebookclave ya que está diseñada específicamente para este paquete.

El siguiente es un ejemplo que muestra todas las etiquetas de campo posibles.

// define a Facebook feed object.
type FacebookFeed struct {
    Id          string            `facebook:",required"`             // this field must exist in response.
                                                                     // mind the "," before "required".
    Story       string
    FeedFrom    *FacebookFeedFrom `facebook:"from"`                  // use customized field name "from".
    CreatedTime string            `facebook:"created_time,required"` // both customized field name and "required" flag.
    Omitted     string            `facebook:"-"`                     // this field is omitted when decoding.
}

type FacebookFeedFrom struct {
    Name string `json:"name"`                   // the "json" key also works as expected.
    Id string   `facebook:"id" json:"shadowed"` // if both "facebook" and "json" key are set, the "facebook" key is used.
}

// create a feed object direct from Graph API result.
var feed FacebookFeed
res, _ := session.Get("/me/feed", nil)
res.DecodeField("data.0", &feed) // read latest feed
Enviar una solicitud por lotes
params1 := Params{
    "method": fb.GET,
    "relative_url": "me",
}
params2 := Params{
    "method": fb.GET,
    "relative_url": uint64(100002828925788),
}
results, err := fb.BatchApi(your_access_token, params1, params2)

if err != nil {
    // check error...
    return
}

// batchResult1 and batchResult2 are response for params1 and params2.
batchResult1, _ := results[0].Batch()
batchResult2, _ := results[1].Batch()

// Use parsed result.
var id string
res := batchResult1.Result
res.DecodeField("id", &id)

// Use response header.
contentType := batchResult1.Header.Get("Content-Type")
Uso con Google App Engine
Google App Engine proporciona el appengine/urlfetchpaquete como el paquete de cliente HTTP estándar. Por esta razón, el cliente predeterminado net/httpno funcionará. Uno debe configurar explícitamente el cliente HTTP Sessionpara que funcione.

import (
    "appengine"
    "appengine/urlfetch"
)

// suppose it's the AppEngine context initialized somewhere.
var context appengine.Context

// default Session object uses http.DefaultClient which is not allowed to use
// in appengine. one has to create a Session and assign it a special client.
seesion := globalApp.Session("a-access-token")
session.HttpClient = urlfetch.Client(context)

// now, the session uses AppEngine HTTP client now.
res, err := session.Get("/me", nil)
Seleccione la versión de Graph API
Consulte Control de versiones de la plataforma para comprender la estrategia de control de versiones de Facebook.

// This package uses the default version which is controlled by the Facebook app setting.
// change following global variable to specify a global default version.
fb.Version = "v3.0"

// starting with Graph API v2.0; it's not allowed to get useful information without an access token.
fb.Api("huan.du", GET, nil)

// it's possible to specify version per session.
session := &fb.Session{}
session.Version = "v3.0" // overwrite global default.
Habilitarappsecret_proof
Facebook puede verificar las llamadas de Graph API con appsecret_proof. Es una función para hacer que las llamadas a Graph API sean más seguras. Consulte Protección de solicitudes de la API Graph para obtener más información al respecto.

globalApp := fb.New("your-app-id", "your-app-secret")

// enable "appsecret_proof" for all sessions created by this app.
globalApp.EnableAppsecretProof = true

// all calls in this session are secured.
session := globalApp.Session("a-valid-access-token")
session.Get("/me", nil)

// it's also possible to enable/disable this feature per session.
session.EnableAppsecretProof(false)
Depuración de solicitudes de API
Facebook ha introducido una forma de depurar las llamadas a la API Graph. Consulte Depuración de solicitudes de API para obtener más detalles.

Este paquete proporciona un nivel de paquete y un indicador de depuración por sesión. Establézcalo Debugen una DEBUG_*constante para cambiar el modo de depuración globalmente o utilícelo Session#SetDebugpara cambiar el modo de depuración para una sesión.

Cuando el modo de depuración está activado, utilícelo Result#DebugInfopara obtener DebugInfola estructura del resultado.

fb.Debug = fb.DEBUG_ALL

res, _ := fb.Get("/me", fb.Params{"access_token": "xxx"})
debugInfo := res.DebugInfo()

fmt.Println("http headers:", debugInfo.Header)
fmt.Println("facebook api version:", debugInfo.FacebookApiVersion)
Supervisión de la información de uso de la API
Llame Result#UsageInfopara obtener una UsageInfoestructura que contenga información del límite de velocidad a nivel de la aplicación y de la página del resultado. Puede encontrar más información sobre la limitación de velocidad aquí .

res, _ := fb.Get("/me", fb.Params{"access_token": "xxx"})
usageInfo := res.UsageInfo()

fmt.Println("App level rate limit information:", usageInfo.App)
fmt.Println("Page level rate limit information:", usageInfo.Page)
fmt.Println("Ad account rate limiting information:", usageInfo.AdAccount)
fmt.Println("Business use case usage information:", usageInfo.BusinessUseCase)
trabajar con paquetegolang.org/x/oauth2
El golang.org/x/oauth2paquete puede manejar bastante bien el proceso de autenticación OAuth2 de Facebook y el token de acceso. Este paquete puede funcionar con él configurando Session#HttpClientel cliente de OAuth2.

import (
    "golang.org/x/oauth2"
    oauth2fb "golang.org/x/oauth2/facebook"
    fb "github.com/huandu/facebook/v2"
)

// Get Facebook access token.
conf := &oauth2.Config{
    ClientID:     "AppId",
    ClientSecret: "AppSecret",
    RedirectURL:  "CallbackURL",
    Scopes:       []string{"email"},
    Endpoint:     oauth2fb.Endpoint,
}
token, err := conf.Exchange(oauth2.NoContext, "code")

// Create a client to manage access token life cycle.
client := conf.Client(oauth2.NoContext, token)

// Use OAuth2 client with session.
session := &fb.Session{
    Version:    "v2.4",
    HttpClient: client,
}

// Use session.
res, _ := session.Get("/me", nil)
Control de tiempo de espera y cancelación conContext
El Sessionaceptar un Context.

// Create a new context.
ctx, cancel := context.WithTimeout(session.Context(), 100 * time.Millisecond)
defer cancel()

// Call an API with ctx.
// The return value of `session.WithContext` is a shadow copy of original session and
// should not be stored. It can be used only once.
result, err := session.WithContext(ctx).Get("/me", nil)
Consulte esta publicación de blog de Go sobre contexto para obtener más detalles sobre cómo usar Context.

Registro de cambios
Consulte CHANGELOG.md .

Fuera del ámbito
Sin integración de OAuth. Este paquete solo proporciona API para analizar/verificar el token de acceso y el código generado en el proceso de autenticación de OAuth 2.0.
No hay API RESTful antigua ni compatibilidad con FQL. Dichas API están en desuso durante años. Olvídate de ellos.
Licencia
Este paquete tiene la licencia MIT. Ver LICENCIA para más detalles.
