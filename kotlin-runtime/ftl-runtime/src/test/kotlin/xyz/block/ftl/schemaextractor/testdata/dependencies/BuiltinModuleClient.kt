// Built-in types for FTL.
//
// This is copied from the FTL runtime and is not meant to be edited.
package ftl.builtin

/**
 * HTTP request structure used for HTTP ingress verbs.
 */
public data class HttpRequest<Body>(
  public val method: String,
  public val path: String,
  public val pathParameters: Map<String, String>,
  public val query: Map<String, ArrayList<String>>,
  public val headers: Map<String, ArrayList<String>>,
  public val body: Body,
)

/**
 * HTTP response structure used for HTTP ingress verbs.
 */
public data class HttpResponse<Body, Error>(
  public val status: Long,
  public val headers: Map<String, ArrayList<String>>,
  public val body: Body? = null,
  public val error: Error? = null,
)

public class Empty
