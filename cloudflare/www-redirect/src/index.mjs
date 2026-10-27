export default {
  fetch(request) {
    const target = new URL(request.url);
    if (target.hostname !== "www.damiengoehrig.ca") {
      return new Response("Not found", { status: 404 });
    }
    target.protocol = "https:";
    target.hostname = "damiengoehrig.ca";
    target.port = "";
    return Response.redirect(target.toString(), 301);
  },
};
