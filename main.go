package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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

func add(query string, args ...string) {
	conn, err := neoPool.OpenPool()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for _, v := range args {
		_, err = conn.ExecNeo(query, map[string]interface{}{"name": v})
		if err != nil {
			if strings.Contains(err.Error(), "Neo.ClientError.Schema.ConstraintValidationFailed") {
				fmt.Println(fmt.Sprintf("constraint violation, no-op on: [%s]", v))
				continue
			}
			panic(err)
		}
	}
}

func list(query string, limit int) []string {
	conn, err := neoPool.OpenPool()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	r, err := conn.QueryNeo(query, map[string]interface{}{"limit": limit})
	if err != nil {
		panic(err)
	}

	list := []string{}

	rows, _, err := r.All()
	if err != nil {
		panic(err)
	}

	for _, n := range rows {
		list = append(list, n[0].(string))
	}

	return list
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
					add("CREATE (s:Skill { name: {name} });", c.Args()...)
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
					add("CREATE (p:Person { name: {name} });", c.Args()...)
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
					add("CREATE (j:Project { name: {name} });", c.Args()...)
					//TODO: Add optional attributes like other names (aliases, check for uniqueness), organization, length, year
				} else {
					fmt.Println("must provide argument")
				}
				return nil
			},
		},
		{
			Name:    "list skills",
			Aliases: []string{"list-skills"},
			Usage:   "list all skills currently being tracked",
			Action: func(c *cli.Context) error {
				limit := "25"
				if c.NArg() != 0 {
					limit = c.Args()[0]
				}

				ln, err := strconv.Atoi(limit)
				if err != nil {
					fmt.Println(limit + " <<-- this is not a number!")
					return nil
				}

				l := list("MATCH (n:Skill ) RETURN n.name as name LIMIT {limit} ;", ln)

				if l != nil {
					fmt.Println("List of skills")
					for i, n := range l {
						i++
						fmt.Println(strconv.Itoa(i) + ": " + n)
					}
				}

				return nil
			},
		},
		{
			Name:    "list people",
			Aliases: []string{"list-people"},
			Usage:   "list all people currently being tracked",
			Action: func(c *cli.Context) error {
				limit := "25"
				if c.NArg() != 0 {
					limit = c.Args()[0]
				}

				ln, err := strconv.Atoi(limit)
				if err != nil {
					fmt.Println(limit + " <<-- this is not a number!")
					return nil
				}

				l := list("MATCH (n:Person ) RETURN n.name as name LIMIT {limit} ;", ln)

				if l != nil {
					fmt.Println("List of people")
					for i, n := range l {
						i++
						fmt.Println(strconv.Itoa(i) + ": " + n)
					}
				}

				return nil
			},
		},
		{
			Name:    "list projects",
			Aliases: []string{"list-projects"},
			Usage:   "list all projects currently being tracked",
			Action: func(c *cli.Context) error {
				limit := "25"
				if c.NArg() != 0 {
					limit = c.Args()[0]
				}

				ln, err := strconv.Atoi(limit)
				if err != nil {
					fmt.Println(limit + " <<-- this is not a number!")
					return nil
				}

				l := list("MATCH (n:Project ) RETURN n.name as name LIMIT {limit} ;", ln)

				if l != nil {
					fmt.Println("List of projects")
					for i, n := range l {
						i++
						fmt.Println(strconv.Itoa(i) + ": " + n)
					}
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
