using System;
using Microsoft.EntityFrameworkCore.Migrations;
using Npgsql.EntityFrameworkCore.PostgreSQL.Metadata;

#nullable disable

namespace QuakeStats.Infrastructure.Migrations
{
    /// <inheritdoc />
    public partial class AddMatches : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.RenameColumn(
                name: "type",
                table: "events",
                newName: "event_type");

            migrationBuilder.RenameColumn(
                name: "data",
                table: "events",
                newName: "event_data");

            migrationBuilder.AddColumn<bool>(
                name: "processed",
                table: "events",
                type: "boolean",
                nullable: false,
                defaultValue: false);

            migrationBuilder.CreateTable(
                name: "matches",
                columns: table => new
                {
                    id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    state = table.Column<int>(type: "integer", nullable: false),
                    match_guid = table.Column<Guid>(type: "uuid", nullable: false),
                    map = table.Column<string>(type: "text", nullable: false),
                    game_type = table.Column<int>(type: "integer", nullable: false),
                    created_at = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("pk_matches", x => x.id);
                });
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "matches");

            migrationBuilder.DropColumn(
                name: "processed",
                table: "events");

            migrationBuilder.RenameColumn(
                name: "event_type",
                table: "events",
                newName: "type");

            migrationBuilder.RenameColumn(
                name: "event_data",
                table: "events",
                newName: "data");
        }
    }
}
