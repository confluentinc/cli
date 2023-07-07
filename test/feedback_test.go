package test

import "fmt"

func (s *CLITestSuite) TestFeedback() {
	feedback := "Lorem ipsum dolor sit amet. Qui amet molestiae eum eaque perferendis et eligendi consequatur in sequi quidem. A officiis facilis ea laborum aspernatur ad quia dolor. Est explicabo voluptatem ea aperiam repellat sit sint doloribus vel sint molestiae et unde iure? Quo quis blanditiis ad unde doloribus nam deleniti magni. Sit delectus sint quo quod illum vel ipsum voluptatem ab fugiat rerum hic rerum provident aut nobis quis. Et pariatur explicabo vel excepturi magnam quo quia accusantium. Eum blanditiis blanditiis cum quia debitis in voluptatem tenetur non adipisci placeat qui omnis quibusdam id minus voluptatum. Ut eaque ratione id omnis quod et voluptatem quaerat qui odio voluptatem sit asperiores aliquam! Sed omnis accusantium vel galisum veniam vel rerum ipsum non eius voluptate. Sit debitis velit hic ullam consequatur in assumenda officiis et totam voluptate eos illo rerum rem consequatur accusamus. Hic illo eius At repellat nesciunt ut rerum nulla aut exercitationem ducimus et nisi galisum eos velit quae et animi eligendi. Lorem ipsum dolor sit amet. Sit eveniet excepturi est earum aperiam eos porro earum est dolorem odit? Qui nihil internos At aliquid expedita ut dicta ullam At omnis quibusdam qui odit temporibus sed consequuntur assumenda sit omnis totam! Eos praesentium distinctio aut consequatur iure qui sunt deleniti et nisi voluptatum qui consequatur tenetur et praesentium error. Quo quidem quia ut corrupti deserunt non provident odit sit quia placeat aut atque odio in vero galisum eum enim quod. Ut tenetur accusamus et dicta voluptatem At voluptatem quaerat et officiis sapiente ut cupiditate odit non repellendus omnis. Ut minus excepturi vel incidunt veritatis et optio odit a dolorum molestiae sit deserunt quidem. Aut voluptates molestiae eos dolor beatae sit reprehenderit iure. Sed atque molestiae ex voluptatem ullam in error maiores. Non dolor deserunt in obcaecati similique quo omnis repellat id debitis provident! Eos sunt aliquam qui aliquid suscipit ut consequuntur reprehenderit. Ut distinctio commodi eum debitis fugit qui dolor corporis aut voluptas minus ad excepturi iste. Non quia molestias ea velit delectus qui nihil natus et reprehenderit veniam hic harum voluptatum et nesciunt impedit qui natus aliquid. Et reiciendis odit qui obcaecati molestiae non sequi minus et laboriosam placeat ea dignissimos rerum non quisquam accusantium hic quia doloremque. Id illum dolor aut rerum magnam eum consequatur deleniti est placeat ducimus quo dolores velit.\n"
	tests := []CLITest{
		{args: "feedback", fixture: "feedback/no-confirm.golden", input: "n\n"},
		{args: "feedback", fixture: "feedback/received.golden", input: "y\nThis CLI is great!\n"},
		{args: "feedback", exitCode: 1, fixture: "feedback/too-long.golden", input: fmt.Sprintf("y\n%s", feedback)},
	}
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
