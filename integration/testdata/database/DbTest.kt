package ftl.dbtest

import xyz.block.ftl.Context
import xyz.block.ftl.Verb
import xyz.block.ftl.Database

data class DbRequest(val data: String?)
data class DbResponse(val message: String? = "ok")

val db = Database("testdb")

class DbTest {
  @Verb
  fun create(context: Context, req: DbRequest): DbResponse {
    persistRequest(req)
    return DbResponse()
  }

  fun persistRequest(req: DbRequest) {
    db.conn {
      it.prepareStatement(
        """
        CREATE TABLE IF NOT EXISTS requests
        (
          data TEXT,
          created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
          updated_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc')
       );
       """
      ).execute()
      it.prepareStatement("INSERT INTO requests (data) VALUES ('${req.data}');")
        .execute()
    }
  }
}
