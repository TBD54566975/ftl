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
      
      let requestData = try {
         if let requestRoot = request.ftlEncode() {
            return try JSONSerialization.data(withJSONObject: requestRoot, options: [.fragmentsAllowed])
         }
         // TODO: come up with a proper way to handle this
         return "null".data(using: .utf8)!
      }()
//      throw FTLError(message:"requestData: \(String(data: requestData, encoding: .utf8) ?? "nil")")
      grpcRequest.body = requestData
      let grpcResponse = try await self.grpcClient.call(grpcRequest)
      if grpcResponse.error.message.lengthOfBytes(using: .utf8) > 0 {
         throw FTLError(message:grpcResponse.error.message)
      }
      if let response = Unit() as? Resp {
         return response
      }
      do {
         let root = try JSONSerialization.jsonObject(with: grpcResponse.body, options:[.fragmentsAllowed])
         return try Resp.ftlDecode(root)
      }
      catch let error {
         throw VerbServiceProviderError(message: "could not parse response: \(error)\n\(String(data:grpcResponse.body, encoding:.utf8))")
      }
   }
}
