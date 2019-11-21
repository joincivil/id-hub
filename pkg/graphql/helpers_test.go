package graphql_test

import (
	"testing"
	"time"

	"github.com/joincivil/id-hub/pkg/graphql"
	"github.com/joincivil/id-hub/pkg/utils"
)

func TestConvertInputPublicKeys(t *testing.T) {
	inputPk1 := &graphql.DidDocPublicKeyInput{
		ID:           utils.StrToPtr("did:ethuri:123456"),
		Controller:   utils.StrToPtr("did:ethuri:123456"),
		Type:         utils.StrToPtr("EcdsaSecp256k1VerificationKey2019"),
		PublicKeyHex: utils.StrToPtr("04debef3fcbef3f5659f9169bad80044b287139a401b5da2979e50b032560ed33927eab43338e9991f31185b3152735e98e0471b76f18897d764b4e4f8a7e8f61b"),
	}
	inputPks := []*graphql.DidDocPublicKeyInput{inputPk1}

	pks, pkMap, err := graphql.ConvertInputPublicKeys(inputPks)
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}
	if len(pks) != 1 {
		t.Errorf("Should have gotten 1 pks")
	}
	if len(pkMap) != 1 {
		t.Errorf("Should have gotten 1 items in the pk map: len: %v", len(pkMap))
	}
}

func TestConvertInputAuthentications(t *testing.T) {
	inputPk1 := &graphql.DidDocPublicKeyInput{
		ID:           utils.StrToPtr("did:ethuri:123456"),
		Controller:   utils.StrToPtr("did:ethuri:123456"),
		Type:         utils.StrToPtr("EcdsaSecp256k1VerificationKey2019"),
		PublicKeyHex: utils.StrToPtr("04debef3fcbef3f5659f9169bad80044b287139a401b5da2979e50b032560ed33927eab43338e9991f31185b3152735e98e0471b76f18897d764b4e4f8a7e8f61b"),
	}
	idOnly := false
	inputAuth1 := &graphql.DidDocAuthenticationInput{
		PublicKey: inputPk1,
		IDOnly:    &idOnly,
	}
	inputAuths := []*graphql.DidDocAuthenticationInput{inputAuth1}

	auths, err := graphql.ConvertInputAuthentications(inputAuths, map[string]int{})
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}
	if len(auths) != 1 {
		t.Errorf("Should have gotten 1 auth")
	}
}

func TestConvertInputServices(t *testing.T) {
	inputSrv1 := &graphql.DidDocServiceInput{
		ID:              utils.StrToPtr("did:ethuri:123456#vcr"),
		Type:            utils.StrToPtr("CredentialRepositoryService"),
		Description:     utils.StrToPtr("This is a description"),
		ServiceEndpoint: &utils.AnyValue{Value: "https://repository.example.com/service/8377464"},
	}
	inputSrvs := []*graphql.DidDocServiceInput{inputSrv1}

	srvs, err := graphql.ConvertInputServices(inputSrvs)
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}
	if len(srvs) != 1 {
		t.Errorf("Should have gotten 1 srv")
	}
}

func TestConvertInputProof(t *testing.T) {
	ts := time.Now()
	inputProof := &graphql.LinkedDataProofInput{
		Type:       utils.StrToPtr("LinkedDataSignature2015"),
		Creator:    utils.StrToPtr("did:ethuri:123456"),
		Created:    &ts,
		ProofValue: utils.StrToPtr("thisisasignature value"),
	}

	_, err := graphql.ConvertInputProof(inputProof)
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}
}

func TestInputClaimToContentCredential(t *testing.T) {
	testInput := `{
		"@context": [
			"https://www.w3.org/2018/credentials/v1",
			"https://id.civil.co/credentials/contentcredential/v1"
		],
		"type": [
			"VerifiableCredential",
			"ContentCredential"
		],
		"credentialSubject": {
			"id": "",
			"metadata": {
				"Title": "Civil unrest haunts those who remember Latin America's juntas",
				"RevisionContentHash": "",
				"RevisionContentURL": "",
				"CanonicalURL": "",
				"Slug": "latam-military",
				"Description": "\n         &lt;p&gt;Unrest in Latin America is putting the region&apos;s soldiers back into the public eye.&lt;/p&gt;\n         &lt;p&gt;Recent events in Bolivia, Ecuador and Chile reflect a broad shift in the region that&apos;s testing the military, police and governments alike. Sick of austerity, feeling left behind, people are venting by taking to the streets, clashing with police and setting buildings on fire. They&apos;ve targeted key national infrastructure including oil fields.&lt;/p&gt;\n         &lt;p&gt;Leaders face movements which have morphed into multiple demands, sometimes from multiple groups. In the case of Bolivia, South America&apos;s longest-serving leader Evo Morales failed to control things at all after a disputed election, and ended up fleeing to Mexico.&lt;/p&gt;\n         &lt;p&gt;These weeks of discontent have focused attention on the military. Army chiefs naturally have vested interests in certain leaders or certain outcomes.&lt;/p&gt;\n         &lt;p&gt;But in a region with dark memories of dictatorships there are also dangers in being drawn in. The violent actions of protesters can see soldiers deployed in situations for which they are unprepared or trained but where they also have superior weaponry to the police, prompting them to overreact.&lt;/p&gt;\n         &lt;p&gt;&quot;The participation of military forces in the control of social unrest has to be the exception, for situations specifically provided for by law and not a rule as they are increasingly happening in Latin America,&quot; said Rocio San Miguel, Venezuela-based president of watchdog group Control Ciudadano.&lt;/p&gt;\n         &lt;p&gt;&quot;In Latin America we live with the ghost of the terrible violations committed to human rights under the framework of national security, in which an internal enemy was established for political reasons.&quot;&lt;/p&gt;\n         &lt;p&gt;For some, the move by the Bolivian army -- its chief called publicly for Morales to step down -- was a reminder of the 1960s through 1980s, when coups led to brutal right-wing dictatorships from Buenos Aires to Brasilia.&lt;/p&gt;\n         &lt;p&gt;And Morales&apos; exit has failed to stem the crisis. Clashes intensified on Friday with local media reporting at least five people were killed by security forces as they protested the interim government of Jeanine Anez.&lt;/p&gt;\n         &lt;p&gt;As of last year, trust in the armed forces across the region remained high. Years of disappointing growth and weariness over endemic corruption has seen confidence in politicians decline. In contrast, almost everywhere the armed forces are among the most respected institutions, behind only the church, according to Latinobarometro, a regional public opinion survey.&lt;/p&gt;\n         &lt;p&gt;During the unrest in Ecuador and Chile, leaders have posted pictures with army officers. In Venezuela, Nicolas Maduro has hung onto power because the opposition has failed to draw his top officers away. And in Peru, President Martin Vizcarra posed with military chiefs while he fended off a challenge from the opposition which saw him dissolve parliament.&lt;/p&gt;\n         &lt;p&gt;The military has also made some effort to move away from the historical narrative of armed interventions.&lt;/p&gt;\n         &lt;p&gt;In recent years it&apos;s often been troops showing up to rescue citizens from natural disasters like mudslides and earthquakes. In countries with large, poor indigenous populations, people identify more with soldiers than wealthier politicians in the major cities.&lt;/p&gt;\n         &lt;p&gt;That&apos;s all the more so as economic growth in Latin America lags developing-nation peers elsewhere.&lt;/p&gt;\n         &lt;p&gt;&quot;Lately the armed forces have been called to maintain law and order, to clean beaches, to help with infrastructure projects,&quot; said retired Brazilian army general Paulo Chagas. &quot;It shows that the armed forces participate and are not simply decorative.&quot;&lt;/p&gt;\n         &lt;p&gt;But at the same time it is not recommended, according to Chagas. &quot;First, because it generates image and resources wear. Second, it shows state structure deficiency.&quot;&lt;/p&gt;\n         &lt;p&gt;Craig Deare, a professor at the National Defense University in Washington, which teaches U.S. and foreign soldiers, said troops were being used for basic police functions, when their training is for national defense and war, because the police were ineffective or corrupt.&lt;/p&gt;\n         &lt;p&gt;&quot;To see a heavy military presence in the streets in some ways is reassuring if you&apos;re concerned about security,&quot; said Deare, who was briefly senior director for Western Hemisphere Affairs for the National Security Council in the Trump administration. &quot;But as you reflect in the longer run, what does that mean for the quality of the political system that it&apos;s not capable of ensuring peace and tranquility? It&apos;s a concern.&quot;&lt;/p&gt;\n         &lt;p&gt;Chile&apos;s President Sebastian Pinera for one is being careful. Even as he grappled with violent protests, he quickly returned the military to barracks after criticism over the deaths of at least 19 people, opting to leave security to an overwhelmed police force. He has though appeared in public next to senior army officials.&lt;/p&gt;\n         &lt;p&gt;In Mexico the army isn&apos;t being used to curb unrest. But it has played a large role for the past decade against drug violent cartels. It is being used to populate the National Guard, President Andres Manuel Lopez Obrador&apos;s new force tasked with pacifying the country, which includes stopping migrants heading to the U.S.&lt;/p&gt;\n         &lt;p&gt;Argentina is an outlier. Its military isn&apos;t popular and is still criticized for its role in the dictatorship from 1976-83 when there were severe human rights abuses.&lt;/p&gt;\n         &lt;p&gt;&quot;Latin America&apos;s 20th century history shows us the consequences of handing over nearly all civilian functions to the military -- gross human rights abuses, reigns of terror, and often widespread unpunished corruption,&quot; said Adam Isacson, director for defense oversight at The Washington Office on Latin America, a group that promotes democracy.&lt;/p&gt;\n         &lt;p&gt;That doesn&apos;t mean the army is about to fade away, though.&lt;/p&gt;\n         &lt;p&gt;In Venezuela, Maduro has given officers control over large swaths of the economy.&lt;/p&gt;\n         &lt;p&gt;Opposition leaders and the U.S. have tried to convince individual commanders to change sides with offers of amnesty. Only a few hundred, mostly lower-level troops, responded. Still, if Maduro were to lose control of the armed forces, he&apos;d struggle to stay in power.&lt;/p&gt;\n         &lt;p&gt;&quot;The role of the armed forces is very similar to that of Cuba, in which it has an important management over the economy, so it has incentives to keep the government in power,&quot; said Diego Moya-Ocampos, a political-risk consultant at IHS Markit in London. &quot;The government, alongside segments of the higher ranks, constantly follows, spies and intimidates officers in the junior ranks when there&apos;s suspicion of insurrection plans.&quot;&lt;/p&gt;\n         &lt;p&gt;And sometimes the influence can be cloaked in the veneer of respectable politics.&lt;/p&gt;\n         &lt;p&gt;In Brazil&apos;s last election the army aligned with former army captain Jair Bolsonaro, who picked General Hamilton Mourao as his vice president.&lt;/p&gt;\n         &lt;p&gt;Still, some senior generals have already left posts in his administration, worried his tumble in popularity will hurt the army&apos;s reputation, said a person familiar with their thinking.&lt;/p&gt;\n         &lt;p&gt;Their thinking is the military can survive different governments, and that it must not be contaminated, the person said.&lt;/p&gt;\n         &lt;p&gt;&quot;It&apos;s a measure of political weakness -- we&apos;re at a moment when leaders in country after country where political institutions are being questioned and not seen as credible by citizens,&quot; said Michael Shifter, president of the Inter-American Dialogue, a think tank, and a professor at Georgetown University.&lt;/p&gt;\n         &lt;p&gt;&quot;It is not a good sign for democracy that the military is the arbiter,&quot; he said. &quot;It&apos;s a reflection of the bankruptcy of political parties and leaders who need to rely on the military to govern and stay in power.&quot;&lt;/p&gt;\n         &lt;p/&gt;\n         &lt;p&gt;- - - &lt;/p&gt;\n         &lt;p/&gt;\n         &lt;p&gt;Bloomberg&apos;s Simone Iglesias, Fabiola Zerpa, John Quigley, Philip Sanders, Eduardo Thomson and Patricia Laya contributed to this report.&lt;/p&gt;\n      \n   ",
				"Contributors": [
					{
						"Role": "author",
						"Name": "Eric Martin"
					},
					{
						"Role": "author",
						"Name": "Patrick Gillespie"
					},
					{
						"Role": "author",
						"Name": "Samy Adghirni"
					}
				],
				"Images": [
					{
						"URL": "https://news-service.s3.amazonaws.com/latam-military-4ab93a04-0a27-11ea-8397-a955cd542d00.jpg",
						"Hash": "",
						"H": 2670,
						"W": 4000
					}
				],
				"Tags": [],
				"PrimaryTag": "",
				"RevisionDate": "0001-01-01T00:00:00Z",
				"OriginalPublishDate": "2019-11-18T17:18:10Z",
				"Opinion": false,
				"CivilSchemaVersion": ""
			}
		},
		"issuer": "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1",
		"credentialSchema": {
			"id": "https://id.civil.co/credentials/schemas/v1/metadata.json",
			"type": "JsonSchemaValidator2018"
		},
		"issuanceDate": "2019-11-18T17:21:09.613186712Z",
		"proof": {
			"type": "EcdsaSecp256k1Signature2019",
			"creator": "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1",
			"created": "2019-11-18T17:21:09.613729268Z",
			"proofValue": "80ed91bd852ba71ef230b74acb66375fe1516c6e282cb202fe10dcf6c0cc14934c179b88a3da16e3737283cd597732936bc51631bdc2596edc2d82d65d9610f900"
		}
	}`

	in := &graphql.ClaimSaveRequestInput{ClaimJSON: utils.StrToPtr(testInput)}
	cc, err := graphql.InputClaimToContentCredential(in)
	if err != nil {
		t.Errorf("Should not have error converting claim to content credential: err: %v", err)
	}
	if cc == nil {
		t.Errorf("Should have returned a non-nil content credential")
	}
	t.Logf("proof type = %T\n", cc.Proof)
}

func TestInputClaimToContentCredentialProofSlice(t *testing.T) {
	testInput := `{
		"@context": [
			"https://www.w3.org/2018/credentials/v1",
			"https://id.civil.co/credentials/contentcredential/v1"
		],
		"type": [
			"VerifiableCredential",
			"ContentCredential"
		],
		"credentialSubject": {
			"id": "",
			"metadata": {
				"Title": "Civil unrest haunts those who remember Latin America's juntas",
				"RevisionContentHash": "",
				"RevisionContentURL": "",
				"CanonicalURL": "",
				"Slug": "latam-military",
				"Description": "\n         &lt;p&gt;Unrest in Latin America is putting the region&apos;s soldiers back into the public eye.&lt;/p&gt;\n         &lt;p&gt;Recent events in Bolivia, Ecuador and Chile reflect a broad shift in the region that&apos;s testing the military, police and governments alike. Sick of austerity, feeling left behind, people are venting by taking to the streets, clashing with police and setting buildings on fire. They&apos;ve targeted key national infrastructure including oil fields.&lt;/p&gt;\n         &lt;p&gt;Leaders face movements which have morphed into multiple demands, sometimes from multiple groups. In the case of Bolivia, South America&apos;s longest-serving leader Evo Morales failed to control things at all after a disputed election, and ended up fleeing to Mexico.&lt;/p&gt;\n         &lt;p&gt;These weeks of discontent have focused attention on the military. Army chiefs naturally have vested interests in certain leaders or certain outcomes.&lt;/p&gt;\n         &lt;p&gt;But in a region with dark memories of dictatorships there are also dangers in being drawn in. The violent actions of protesters can see soldiers deployed in situations for which they are unprepared or trained but where they also have superior weaponry to the police, prompting them to overreact.&lt;/p&gt;\n         &lt;p&gt;&quot;The participation of military forces in the control of social unrest has to be the exception, for situations specifically provided for by law and not a rule as they are increasingly happening in Latin America,&quot; said Rocio San Miguel, Venezuela-based president of watchdog group Control Ciudadano.&lt;/p&gt;\n         &lt;p&gt;&quot;In Latin America we live with the ghost of the terrible violations committed to human rights under the framework of national security, in which an internal enemy was established for political reasons.&quot;&lt;/p&gt;\n         &lt;p&gt;For some, the move by the Bolivian army -- its chief called publicly for Morales to step down -- was a reminder of the 1960s through 1980s, when coups led to brutal right-wing dictatorships from Buenos Aires to Brasilia.&lt;/p&gt;\n         &lt;p&gt;And Morales&apos; exit has failed to stem the crisis. Clashes intensified on Friday with local media reporting at least five people were killed by security forces as they protested the interim government of Jeanine Anez.&lt;/p&gt;\n         &lt;p&gt;As of last year, trust in the armed forces across the region remained high. Years of disappointing growth and weariness over endemic corruption has seen confidence in politicians decline. In contrast, almost everywhere the armed forces are among the most respected institutions, behind only the church, according to Latinobarometro, a regional public opinion survey.&lt;/p&gt;\n         &lt;p&gt;During the unrest in Ecuador and Chile, leaders have posted pictures with army officers. In Venezuela, Nicolas Maduro has hung onto power because the opposition has failed to draw his top officers away. And in Peru, President Martin Vizcarra posed with military chiefs while he fended off a challenge from the opposition which saw him dissolve parliament.&lt;/p&gt;\n         &lt;p&gt;The military has also made some effort to move away from the historical narrative of armed interventions.&lt;/p&gt;\n         &lt;p&gt;In recent years it&apos;s often been troops showing up to rescue citizens from natural disasters like mudslides and earthquakes. In countries with large, poor indigenous populations, people identify more with soldiers than wealthier politicians in the major cities.&lt;/p&gt;\n         &lt;p&gt;That&apos;s all the more so as economic growth in Latin America lags developing-nation peers elsewhere.&lt;/p&gt;\n         &lt;p&gt;&quot;Lately the armed forces have been called to maintain law and order, to clean beaches, to help with infrastructure projects,&quot; said retired Brazilian army general Paulo Chagas. &quot;It shows that the armed forces participate and are not simply decorative.&quot;&lt;/p&gt;\n         &lt;p&gt;But at the same time it is not recommended, according to Chagas. &quot;First, because it generates image and resources wear. Second, it shows state structure deficiency.&quot;&lt;/p&gt;\n         &lt;p&gt;Craig Deare, a professor at the National Defense University in Washington, which teaches U.S. and foreign soldiers, said troops were being used for basic police functions, when their training is for national defense and war, because the police were ineffective or corrupt.&lt;/p&gt;\n         &lt;p&gt;&quot;To see a heavy military presence in the streets in some ways is reassuring if you&apos;re concerned about security,&quot; said Deare, who was briefly senior director for Western Hemisphere Affairs for the National Security Council in the Trump administration. &quot;But as you reflect in the longer run, what does that mean for the quality of the political system that it&apos;s not capable of ensuring peace and tranquility? It&apos;s a concern.&quot;&lt;/p&gt;\n         &lt;p&gt;Chile&apos;s President Sebastian Pinera for one is being careful. Even as he grappled with violent protests, he quickly returned the military to barracks after criticism over the deaths of at least 19 people, opting to leave security to an overwhelmed police force. He has though appeared in public next to senior army officials.&lt;/p&gt;\n         &lt;p&gt;In Mexico the army isn&apos;t being used to curb unrest. But it has played a large role for the past decade against drug violent cartels. It is being used to populate the National Guard, President Andres Manuel Lopez Obrador&apos;s new force tasked with pacifying the country, which includes stopping migrants heading to the U.S.&lt;/p&gt;\n         &lt;p&gt;Argentina is an outlier. Its military isn&apos;t popular and is still criticized for its role in the dictatorship from 1976-83 when there were severe human rights abuses.&lt;/p&gt;\n         &lt;p&gt;&quot;Latin America&apos;s 20th century history shows us the consequences of handing over nearly all civilian functions to the military -- gross human rights abuses, reigns of terror, and often widespread unpunished corruption,&quot; said Adam Isacson, director for defense oversight at The Washington Office on Latin America, a group that promotes democracy.&lt;/p&gt;\n         &lt;p&gt;That doesn&apos;t mean the army is about to fade away, though.&lt;/p&gt;\n         &lt;p&gt;In Venezuela, Maduro has given officers control over large swaths of the economy.&lt;/p&gt;\n         &lt;p&gt;Opposition leaders and the U.S. have tried to convince individual commanders to change sides with offers of amnesty. Only a few hundred, mostly lower-level troops, responded. Still, if Maduro were to lose control of the armed forces, he&apos;d struggle to stay in power.&lt;/p&gt;\n         &lt;p&gt;&quot;The role of the armed forces is very similar to that of Cuba, in which it has an important management over the economy, so it has incentives to keep the government in power,&quot; said Diego Moya-Ocampos, a political-risk consultant at IHS Markit in London. &quot;The government, alongside segments of the higher ranks, constantly follows, spies and intimidates officers in the junior ranks when there&apos;s suspicion of insurrection plans.&quot;&lt;/p&gt;\n         &lt;p&gt;And sometimes the influence can be cloaked in the veneer of respectable politics.&lt;/p&gt;\n         &lt;p&gt;In Brazil&apos;s last election the army aligned with former army captain Jair Bolsonaro, who picked General Hamilton Mourao as his vice president.&lt;/p&gt;\n         &lt;p&gt;Still, some senior generals have already left posts in his administration, worried his tumble in popularity will hurt the army&apos;s reputation, said a person familiar with their thinking.&lt;/p&gt;\n         &lt;p&gt;Their thinking is the military can survive different governments, and that it must not be contaminated, the person said.&lt;/p&gt;\n         &lt;p&gt;&quot;It&apos;s a measure of political weakness -- we&apos;re at a moment when leaders in country after country where political institutions are being questioned and not seen as credible by citizens,&quot; said Michael Shifter, president of the Inter-American Dialogue, a think tank, and a professor at Georgetown University.&lt;/p&gt;\n         &lt;p&gt;&quot;It is not a good sign for democracy that the military is the arbiter,&quot; he said. &quot;It&apos;s a reflection of the bankruptcy of political parties and leaders who need to rely on the military to govern and stay in power.&quot;&lt;/p&gt;\n         &lt;p/&gt;\n         &lt;p&gt;- - - &lt;/p&gt;\n         &lt;p/&gt;\n         &lt;p&gt;Bloomberg&apos;s Simone Iglesias, Fabiola Zerpa, John Quigley, Philip Sanders, Eduardo Thomson and Patricia Laya contributed to this report.&lt;/p&gt;\n      \n   ",
				"Contributors": [
					{
						"Role": "author",
						"Name": "Eric Martin"
					},
					{
						"Role": "author",
						"Name": "Patrick Gillespie"
					},
					{
						"Role": "author",
						"Name": "Samy Adghirni"
					}
				],
				"Images": [
					{
						"URL": "https://news-service.s3.amazonaws.com/latam-military-4ab93a04-0a27-11ea-8397-a955cd542d00.jpg",
						"Hash": "",
						"H": 2670,
						"W": 4000
					}
				],
				"Tags": [],
				"PrimaryTag": "",
				"RevisionDate": "0001-01-01T00:00:00Z",
				"OriginalPublishDate": "2019-11-18T17:18:10Z",
				"Opinion": false,
				"CivilSchemaVersion": ""
			}
		},
		"issuer": "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1",
		"credentialSchema": {
			"id": "https://id.civil.co/credentials/schemas/v1/metadata.json",
			"type": "JsonSchemaValidator2018"
		},
		"issuanceDate": "2019-11-18T17:21:09.613186712Z",
		"proof": [{
			"type": "EcdsaSecp256k1Signature2019",
			"creator": "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1",
			"created": "2019-11-18T17:21:09.613729268Z",
			"proofValue": "80ed91bd852ba71ef230b74acb66375fe1516c6e282cb202fe10dcf6c0cc14934c179b88a3da16e3737283cd597732936bc51631bdc2596edc2d82d65d9610f900"
		}]
	}`

	in := &graphql.ClaimSaveRequestInput{ClaimJSON: utils.StrToPtr(testInput)}
	cc, err := graphql.InputClaimToContentCredential(in)
	if err != nil {
		t.Errorf("Should not have error converting claim to content credential: err: %v", err)
	}
	if cc == nil {
		t.Errorf("Should have returned a non-nil content credential")
	}
	t.Logf("proof type = %T\n", cc.Proof)
}
