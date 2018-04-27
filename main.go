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

var plurals map[string]string

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
	plurals = map[string]string{
		"skill":   "skills",
		"person":  "people",
		"project": "projects",
	}

	prompt()
}

func add(query string) {
	conn, err := neoPool.OpenPool()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.ExecNeo(query, nil)
	if err != nil {
		if strings.Contains(err.Error(), "Neo.ClientError.Schema.ConstraintValidationFailed") {
			fmt.Println(fmt.Sprintf("constraint violation, no-op on: [%s]", query))
			return
		}
		panic(err)
	}

}

func list(query string) []string {
	conn, err := neoPool.OpenPool()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Running query: " + query)
	r, err := conn.QueryNeo(query, nil)
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

//an alternative/enchancement to this might be to extend each add step to
//list another category and allow users to respond with which items need relationships
func buildRelationship(names map[string]string) {
	var typeX, x, typeY, y string
	for k, v := range names {
		if len(typeX) == 0 {
			typeX = k
			x = v
		} else {
			typeY = k
			y = v
		}
	}

	runQuery(fmt.Sprintf("MATCH (e:%s) WHERE e.name = '%s' MATCH (s:%s) WHERE s.name = '%s' CREATE (e)-[:KNOWS]->(s);", typeX, x, typeY, y))
	//TODO: relationships should be more descriptive/clear - case by case work out which things are being connected and what that means
	//TODO: [:KNOWS {level: "advanced"}] - attributes on the relationship
}

func show(names map[string]string) map[string]map[string][]string {
	conn, err := neoPool.OpenPool()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	results := make(map[string]map[string][]string)

	for t, n := range names {
		q := fmt.Sprintf("MATCH (n:%s {name: '%s'})-[]-(b) RETURN labels(b) AS label, b.name AS name;", t, n)

		r, err := conn.QueryNeo(q, nil)
		if err != nil {
			panic(err)
		}

		forType := make(map[string][]string)

		rows, _, err := r.All()
		if err != nil {
			panic(err)
		}

		for _, n := range rows {

			labels := n[0].([]interface{})
			var labelList []string
			for _, l := range labels {
				labelList = append(labelList, l.(string))
			}
			label := labelList[0]
			name := n[1].(string)

			l := forType[label]
			l = append(l, name)

			forType[label] = l
		}

		results[t] = forType
	}

	return results
}

func prompt() {
	app := cli.NewApp()
	app.Name = "skills"
	app.Usage = "fight the loneliness!"
	app.Action = func(c *cli.Context) error {
		// this could print the existing graph or list the skills/people/projects
		return nil
	}

	var skillFlag, personFlag, projectFlag string

	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"add", "a"},
			Usage:   "add something not currently being tracked",
			Action: func(c *cli.Context) error {
				if c.NArg() == 2 {
					q := fmt.Sprintf("CREATE (j:%s { name: '{%s}' });", strings.Title(c.Args()[0]), c.Args()[1])
					add(q)
					fmt.Println("added: " + c.Args()[1])
					//TODO: Add optional attributes like other names (aliases, check for uniqueness), organization, length, year
				} else {
					fmt.Println("must provide argument")
				}
				return nil
			},
		},
		{
			Name:    "list",
			Aliases: []string{"list", "l"},
			Usage:   "list all elements currently being tracked",
			Action: func(c *cli.Context) error {
				var flag, limit string

				if c.NArg() > 0 && c.NArg() < 3 {
					flag = strings.Title(c.Args()[0])

					limit = "25"
					if c.NArg() == 2 {
						limit = c.Args()[1]
					}
				} else {
					fmt.Println("must provide a type argument (optional limit argument). further arguments are invalid")
					return nil
				}

				l := list(fmt.Sprintf("MATCH (n:%s) RETURN n.name as name LIMIT %s ;", flag, limit))

				if l != nil {
					fmt.Println("List of " + plurals[flag])
					for i, n := range l {
						i++
						fmt.Println(strconv.Itoa(i) + ": " + n)
					}
				}

				return nil
			},
		},
		{
			Name:    "build relationships",
			Aliases: []string{"build", "relate"},
			Usage:   "build a relationship directly between a skill, person or project",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "skill",
					Destination: &skillFlag,
				},
				cli.StringFlag{
					Name:        "person",
					Destination: &personFlag,
				},
				cli.StringFlag{
					Name:        "project",
					Destination: &projectFlag,
				},
			},
			Action: func(c *cli.Context) error {
				var count int
				names := make(map[string]string)
				if len(skillFlag) != 0 {
					count++
					names["Skill"] = skillFlag
					fmt.Println("skill provided: " + skillFlag)
				}
				if len(personFlag) != 0 {
					count++
					names["Person"] = personFlag
					fmt.Println("person provided: " + personFlag)
				}
				if len(projectFlag) != 0 {
					count++
					names["Project"] = projectFlag
					fmt.Println("project provided: " + projectFlag)
				}

				if count == 2 {
					buildRelationship(names)
				} else {
					fmt.Println("can only build relationships between 2 value of different categories")
				}
				return nil
			},
		},
		{
			Name:    "show relationships",
			Aliases: []string{"show"},
			Usage:   "show all relationships associated with a skill, person or project",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "skill",
					Destination: &skillFlag,
				},
				cli.StringFlag{
					Name:        "person",
					Destination: &personFlag,
				},
				cli.StringFlag{
					Name:        "project",
					Destination: &projectFlag,
				},
			},
			Action: func(c *cli.Context) error {
				names := make(map[string]string)
				if len(skillFlag) != 0 {
					names["Skill"] = skillFlag
					fmt.Println("skill provided: " + skillFlag)
				}
				if len(personFlag) != 0 {
					names["Person"] = personFlag
					fmt.Println("person provided: " + personFlag)
				}
				if len(projectFlag) != 0 {
					names["Project"] = projectFlag
					fmt.Println("project provided: " + projectFlag)
				}

				results := show(names)

				//DO THE OUTPUT
				if results != nil {
					for t, r := range results {
						fmt.Println("For the requested: " + t)

						for typ, list := range r {
							fmt.Println("List of related " + typ)

							for i, n := range list {
								i++
								fmt.Println(strconv.Itoa(i) + ": " + n)
							}
						}
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
