using QuakeStats.Api.Configure;

var builder = WebApplication.CreateBuilder(args);
builder.AddServices();

var app = builder.Build();
await app.ConfigureAsync();

await app.RunAsync();
