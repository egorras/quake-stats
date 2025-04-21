using Microsoft.EntityFrameworkCore;
using QuakeStats.Domain.Services;
using QuakeStats.Infrastructure.Data;
using QuakeStats.Infrastructure.Services;
using QuakeStats.Infrastructure.Services.EventHandlers;

namespace QuakeStats.Api.Configure;

public static class ConfigureServices
{
    public static WebApplicationBuilder AddServices(this WebApplicationBuilder builder)
    {
        // Add services to the container.
        // Learn more about configuring OpenAPI at https://aka.ms/aspnet/openapi
        builder.Services.AddOpenApi();

        // Add DbContext with PostgreSQL and snake_case naming convention
        builder.Services.AddDbContext<ApplicationDbContext>(options =>
            options.UseNpgsql(builder.Configuration.GetConnectionString("DefaultConnection"))
                   .UseSnakeCaseNamingConvention()); // This uses the EFCore.NamingConventions package

        // Register event handlers
        builder.Services.AddScoped<IEventHandler, MatchStartedEventHandler>();
        builder.Services.AddScoped<IEventHandler, MatchReportEventHandler>();
        builder.Services.AddScoped<IEventHandler, PlayerConnectEventHandler>();
        builder.Services.AddScoped<IEventHandler, PlayerDisconnectEventHandler>();

        // Register event processor and background service
        builder.Services.AddScoped<IEventProcessor, EventProcessor>();
        builder.Services.AddHostedService<EventProcessingService>();
        return builder;
    }
}
