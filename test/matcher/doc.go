/*
Package matcher implements Gomega matchers for lxkns information model
artifacts, such as containers and container groups (pods in particular). These
matchers can be used in unit tests of applications using the lxkns API and
information model.

While these matchers almost always can be easily written using the existing
Gomega matchers these new matchers offer to spell expectations out in a
domain-specific language.

For instance, instead of:

	Expect(actual).To(ContainElement(And(HaveField("Name", "foo"), HaveField("Type", "docker.com"))))

more succinctly:

	Expect(actual).To(ContainElement(BeADockerContainer(WithName("foo"))))
*/
package matcher
