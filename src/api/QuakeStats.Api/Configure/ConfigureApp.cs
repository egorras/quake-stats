using Microsoft.EntityFrameworkCore;
using QuakeStats.Infrastructure.Data;
using Scalar.AspNetCore;

namespace QuakeStats.Api.Configure;

public static class ConfigureApp
{
    public static async Task<WebApplication> ConfigureAsync(this WebApplication app)
    {
        // Apply migrations at startup
        using (var scope = app.Services.CreateScope())
        {
            var dbContext = scope.ServiceProvider.GetRequiredService<ApplicationDbContext>();
            await dbContext.Database.MigrateAsync();
        }

        // Configure the HTTP request pipeline.
        if (app.Environment.IsDevelopment())
        {
            app.MapOpenApi();
            app.MapScalarApiReference();
        }

        app.UseHttpsRedirection();
        return app;
    }
}
