package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	bolt "github.com/ONSdigital/golang-neo4j-bolt-driver"
	"github.com/urfave/cli"
)

var neoPool bolt.ClosableDriverPool

func main() {
	var err error
	neoPool, err = bolt.NewClosableDriverPool("bolt://localhost:7687", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println("creating constraints...")
	runQuery("CREATE CONSTRAINT ON (n:Skill) ASSERT n.name IS UNIQUE;",
		"CREATE CONSTRAINT ON (f:Project) ASSERT f.name IS UNIQUE;",
		"CREATE CONSTRAINT ON (n:Person) ASSERT n.name IS UNIQUE;")
	fmt.Println("constraints done, obeying prompt...")
	prompt()
}

func run(query string, placeholder string, args ...string) {
	conn, err := neoPool.OpenPool()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for _, v := range args {

		_, err = conn.ExecNeo(query, map[string]interface{}{placeholder: v})
		if err != nil {
			if strings.Contains(err.Error(), "Neo.ClientError.Schema.ConstraintValidationFailed") {
				fmt.Println(fmt.Sprintf("constraint violation, no-op on: [%s]", v))
				continue
			}
			panic(err)
		}
	}
}

func runQuery(query ...string) {
	conn, err := neoPool.OpenPool()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for _, q := range query {
		if _, err = conn.ExecNeo(q, nil); err != nil {
			if strings.Contains(err.Error(), "Neo.ClientError.Schema.ConstraintValidationFailed") {
				fmt.Println(fmt.Sprintf("constraint violation, no-op on: [%s]", q))
				continue
			}
			panic(err)
		}
	}
}

func buildRelationship(x, y string) {

}

func prompt() {
	app := cli.NewApp()
	app.Name = "skills"
	app.Usage = "fight the loneliness!"
	app.Action = func(c *cli.Context) error {
		// this could print the existing graph or list the skills/people/projects
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:    "add skill",
			Aliases: []string{"skill"},
			Usage:   "add a skill not currently being tracked",
			Action: func(c *cli.Context) error {
				if c.NArg() > 0 {
					run("CREATE (s:Skill { name: {name} });", "name", c.Args()...)
				} else {
					fmt.Println("must provide argument")
				}
				return nil
			},
		},
		{
			Name:    "add person",
			Aliases: []string{"person"},
			Usage:   "add a person not currently being tracked",
			Action: func(c *cli.Context) error {
				if c.NArg() > 0 {
					run("CREATE (p:Person { name: {name} });", "name", c.Args()...)
				} else {
					fmt.Println("must provide argument")
				}
				return nil
			},
		},
		{
			Name:    "add project",
			Aliases: []string{"project"},
			Usage:   "add a project not currently being tracked",
			Action: func(c *cli.Context) error {
				if c.NArg() > 0 {
					run("CREATE (j:Project { name: {name} });", "name", c.Args()...)
					//TODO: Add optional attributes like other names (aliases, check for uniqueness), organization, length, year
				} else {
					fmt.Println("must provide argument")
				}
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
