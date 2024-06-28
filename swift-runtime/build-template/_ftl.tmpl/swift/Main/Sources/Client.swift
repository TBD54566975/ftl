import Foundation
import FTL
import GRPC
import NIOCore
import NIOPosix

class Client: FTLClient {
   let channel: any GRPCChannel
   let grpcClient: Xyz_Block_Ftl_V1_VerbServiceAsyncClient
   init(group:MultiThreadedEventLoopGroup, host:String, port:Int) throws {
      // Configure the channel, we're not using TLS so the connection is `insecure`.
      self.channel = try GRPCChannelPool.with(
         target: .host(host, port: port),
         transportSecurity: .plaintext,
         eventLoopGroup: group
      )
      self.grpcClient = Xyz_Block_Ftl_V1_VerbServiceAsyncClient(channel: channel)
   }
   
   deinit {
      try! self.channel.close().wait()
   }
   
   func call<Req, Resp>(module: String, verb: String, request: Req) async throws -> Resp where Req : FTL.FTLType, Resp : FTL.FTLType {
      var grpcRequest = Xyz_Block_Ftl_V1_CallRequest()
      grpcRequest.verb = Xyz_Block_Ftl_V1_Schema_Ref()
      grpcRequest.verb.module = module
      grpcRequest.verb.name = verb
      
      guard let requestRoot = request.ftlEncode() else {
         throw FTLError(message:"expected request to have non-nil json")
      }
      let requestData = try JSONSerialization.data(withJSONObject: requestRoot, options: [])
      grpcRequest.body = requestData
      let grpcResponse = try await self.grpcClient.call(grpcRequest)
      if grpcResponse.error.message.lengthOfBytes(using: .utf8) > 0 {
         throw FTLError(message:grpcResponse.error.message)
      }
      let root = try JSONSerialization.jsonObject(with: grpcResponse.body, options:[])
      return try Resp.ftlDecode(root)
   }
}
