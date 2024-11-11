from dataclasses import dataclass

from ftl import verb


@dataclass
class EchoRequest:
    name: str


@dataclass
class EchoResponse:
    message: str


@verb
def echo(req: EchoRequest) -> EchoResponse:
    return EchoResponse(message=f"ayooo, {req.name}!")
