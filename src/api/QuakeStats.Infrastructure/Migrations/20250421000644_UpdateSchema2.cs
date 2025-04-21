using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace QuakeStats.Infrastructure.Migrations;

/// <inheritdoc />
public partial class UpdateSchema2 : Migration
{
    /// <inheritdoc />
    protected override void Up(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.AlterColumn<DateTimeOffset>(
            name: "created_at",
            table: "players",
            type: "timestamp with time zone",
            nullable: false,
            defaultValueSql: "CURRENT_TIMESTAMP",
            oldClrType: typeof(DateTime),
            oldType: "timestamp with time zone");

        migrationBuilder.AlterColumn<DateTimeOffset>(
            name: "created_at",
            table: "matches",
            type: "timestamp with time zone",
            nullable: false,
            defaultValueSql: "CURRENT_TIMESTAMP",
            oldClrType: typeof(DateTime),
            oldType: "timestamp with time zone");
    }

    /// <inheritdoc />
    protected override void Down(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.AlterColumn<DateTime>(
            name: "created_at",
            table: "players",
            type: "timestamp with time zone",
            nullable: false,
            oldClrType: typeof(DateTimeOffset),
            oldType: "timestamp with time zone",
            oldDefaultValueSql: "CURRENT_TIMESTAMP");

        migrationBuilder.AlterColumn<DateTime>(
            name: "created_at",
            table: "matches",
            type: "timestamp with time zone",
            nullable: false,
            oldClrType: typeof(DateTimeOffset),
            oldType: "timestamp with time zone",
            oldDefaultValueSql: "CURRENT_TIMESTAMP");
    }
}
