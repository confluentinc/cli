package lduser

import "gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

// NewUser creates a new user identified by the given key.
func NewUser(key string) User {
	return User{key: key}
}

// NewAnonymousUser creates a new anonymous user identified by the given key.
func NewAnonymousUser(key string) User {
	return User{key: key, anonymous: ldvalue.NewOptionalBool(true)}
}

// UserBuilder is a mutable object that uses the Builder pattern to specify properties for a User.
// This is the preferred method for constructing a User; direct access to User fields will be
// removed in a future version.
//
// Obtain an instance of UserBuilder by calling NewUserBuilder, then call setter methods such as
// Name to specify any additional user properties, then call Build() to construct the User. All of
// the UserBuilder setters return a reference the same builder, so they can be chained together:
//
//     user := NewUserBuilder("user-key").Name("Bob").Email("test@example.com").Build()
//
// Setters for user attributes that can be designated private return the type
// UserBuilderCanMakeAttributePrivate, so you can chain the AsPrivateAttribute method:
//
//     user := NewUserBuilder("user-key").Name("Bob").AsPrivateAttribute().Build() // Name is now private
//
// A UserBuilder should not be accessed by multiple goroutines at once.
//
// This is defined as an interface rather than a concrete type only for syntactic convenience (see
// UserBuilderCanMakeAttributePrivate). Applications should not implement this interface since the package
// may add methods to it in the future.
type UserBuilder interface {
	// Key changes the unique key for the user being built.
	Key(value string) UserBuilder

	// Secondary sets the secondary key attribute for the user being built.
	//
	// This affects feature flag targeting
	// (https://docs.launchdarkly.com/home/flags/targeting-users#targeting-rules-based-on-user-attributes)
	// as follows: if you have chosen to bucket users by a specific attribute, the secondary key (if set)
	// is used to further distinguish between users who are otherwise identical according to that attribute.
	Secondary(value string) UserBuilderCanMakeAttributePrivate

	// IP sets the IP address attribute for the user being built.
	IP(value string) UserBuilderCanMakeAttributePrivate

	// Country sets the country attribute for the user being built.
	Country(value string) UserBuilderCanMakeAttributePrivate

	// Email sets the email attribute for the user being built.
	Email(value string) UserBuilderCanMakeAttributePrivate

	// FirstName sets the first name attribute for the user being built.
	FirstName(value string) UserBuilderCanMakeAttributePrivate

	// LastName sets the last name attribute for the user being built.
	LastName(value string) UserBuilderCanMakeAttributePrivate

	// Avatar sets the avatar URL attribute for the user being built.
	Avatar(value string) UserBuilderCanMakeAttributePrivate

	// Name sets the full name attribute for the user being built.
	Name(value string) UserBuilderCanMakeAttributePrivate

	// Anonymous sets the anonymous attribute for the user being built.
	//
	// If a user is anonymous, the user key will not appear on your LaunchDarkly dashboard.
	Anonymous(value bool) UserBuilder

	// Custom sets a custom attribute for the user being built.
	//
	//     user := NewUserBuilder("user-key").
	//         Custom("custom-attr-name", ldvalue.String("some-string-value")).AsPrivateAttribute().
	//         Build()
	Custom(attribute string, value ldvalue.Value) UserBuilderCanMakeAttributePrivate

	// CustomAll sets all of the user's custom attributes at once from a ValueMap.
	//
	// UserBuilder has copy-on-write behavior to make this method efficient: if you do not make any
	// changes to custom attributes after this, it reuses the original map rather than allocating a
	// new one.
	CustomAll(ldvalue.ValueMap) UserBuilderCanMakeAttributePrivate

	// SetAttribute sets any attribute of the user being built, specified as a UserAttribute, to a value
	// of type ldvalue.Value.
	//
	// This method corresponds to the GetAttribute method of User. It is intended for cases where user
	// properties are being constructed generically, such as from a list of key-value pairs. Since not
	// all attributes have the same semantics, its behavior is as follows:
	//
	// 1. For built-in attributes, if the value is not of a type that is supported for that attribute,
	// the method has no effect. For Key, the only supported type is string; for Anonymous, the
	// supported types are boolean or null; and for all other built-ins, the supported types are
	// string or null. Custom attributes may be of any type.
	//
	// 2. Setting an attribute to null (ldvalue.Null() or ldvalue.Value{}) is the same as the attribute
	// not being set in the first place.
	//
	// 3. The method always returns the type UserBuilderCanMakeAttributePrivate, so that you can make
	// the attribute private if that is appropriate by calling AsPrivateAttribute(). For attributes
	// that cannot be made private (Key and Anonymous), calling AsPrivateAttribute() on this return
	// value will have no effect.
	SetAttribute(attribute UserAttribute, value ldvalue.Value) UserBuilderCanMakeAttributePrivate

	// Build creates a User from the current UserBuilder properties.
	//
	// The User is independent of the UserBuilder once you have called Build(); modifying the UserBuilder
	// will not affect an already-created User.
	Build() User
}

// UserBuilderCanMakeAttributePrivate is an extension of UserBuilder that allows attributes to be
// made private via the AsPrivateAttribute() method. All UserBuilderCanMakeAttributePrivate setter
// methods are the same as UserBuilder, and apply to the original builder.
//
// UserBuilder setter methods for attributes that can be made private always return this interface.
// See AsPrivateAttribute for details.
type UserBuilderCanMakeAttributePrivate interface {
	UserBuilder

	// AsPrivateAttribute marks the last attribute that was set on this builder as being a private
	// attribute: that is, its value will not be sent to LaunchDarkly.
	//
	// This action only affects analytics events that are generated by this particular user object. To
	// mark some (or all) user attributes as private for all users, use the Config properties
	// PrivateAttributeName and AllAttributesPrivate.
	//
	// Most attributes can be made private, but Key and Anonymous cannot. This is enforced by the
	// compiler, since the builder methods for attributes that can be made private are the only ones
	// that return UserBuilderCanMakeAttributePrivate; therefore, you cannot write an expression like
	// NewUserBuilder("user-key").AsPrivateAttribute().
	//
	// In this example, FirstName and LastName are marked as private, but Country is not:
	//
	//     user := NewUserBuilder("user-key").
	//         FirstName("Pierre").AsPrivateAttribute().
	//         LastName("Menard").AsPrivateAttribute().
	//         Country("ES").
	//         Build()
	AsPrivateAttribute() UserBuilder

	// AsNonPrivateAttribute marks the last attribute that was set on this builder as not being a
	// private attribute: that is, its value will be sent to LaunchDarkly and can appear on the dashboard.
	//
	// This is the opposite of AsPrivateAttribute(), and has no effect unless you have previously called
	// AsPrivateAttribute() for the same attribute on the same user builder. For more details, see
	// AsPrivateAttribute().
	AsNonPrivateAttribute() UserBuilder
}

type userBuilderImpl struct {
	key                         string
	secondary                   ldvalue.OptionalString
	ip                          ldvalue.OptionalString
	country                     ldvalue.OptionalString
	email                       ldvalue.OptionalString
	firstName                   ldvalue.OptionalString
	lastName                    ldvalue.OptionalString
	avatar                      ldvalue.OptionalString
	name                        ldvalue.OptionalString
	anonymous                   ldvalue.OptionalBool
	custom                      ldvalue.ValueMapBuilder
	privateAttrs                map[UserAttribute]struct{}
	privateAttrsCopyOnWrite     bool
	lastAttributeCanMakePrivate UserAttribute
}

// NewUserBuilder constructs a new UserBuilder, specifying the user key.
//
// For authenticated users, the key may be a username or e-mail address. For anonymous users,
// this could be an IP address or session ID.
func NewUserBuilder(key string) UserBuilder {
	return &userBuilderImpl{key: key}
}

// NewUserBuilderFromUser constructs a new UserBuilder, copying all attributes from an existing user. You may
// then call setter methods on the new UserBuilder to modify those attributes.
//
// Custom attributes, and the set of attribute names that are private, are implemented internally as maps.
// Since the User struct does not expose these maps, they are in effect immutable and will be reused from the
// original User rather than copied whenever possible. The UserBuilder has copy-on-write behavior so that it
// only makes copies of these data structures if you actually modify them.
func NewUserBuilderFromUser(fromUser User) UserBuilder {
	builder := &userBuilderImpl{
		key:                     fromUser.key,
		secondary:               fromUser.secondary,
		ip:                      fromUser.ip,
		country:                 fromUser.country,
		email:                   fromUser.email,
		firstName:               fromUser.firstName,
		lastName:                fromUser.lastName,
		avatar:                  fromUser.avatar,
		name:                    fromUser.name,
		anonymous:               fromUser.anonymous,
		privateAttrs:            fromUser.privateAttributes,
		privateAttrsCopyOnWrite: true,
	}
	if fromUser.custom.Count() > 0 {
		builder.custom = ldvalue.ValueMapBuildFromMap(fromUser.custom)
	}
	return builder
}

func (b *userBuilderImpl) canMakeAttributePrivate(attribute UserAttribute) UserBuilderCanMakeAttributePrivate {
	b.lastAttributeCanMakePrivate = attribute
	return b
}

func (b *userBuilderImpl) Key(value string) UserBuilder {
	b.key = value
	return b
}

func (b *userBuilderImpl) Secondary(value string) UserBuilderCanMakeAttributePrivate {
	b.secondary = ldvalue.NewOptionalString(value)
	return b.canMakeAttributePrivate(SecondaryKeyAttribute)
}

func (b *userBuilderImpl) IP(value string) UserBuilderCanMakeAttributePrivate {
	b.ip = ldvalue.NewOptionalString(value)
	return b.canMakeAttributePrivate(IPAttribute)
}

func (b *userBuilderImpl) Country(value string) UserBuilderCanMakeAttributePrivate {
	b.country = ldvalue.NewOptionalString(value)
	return b.canMakeAttributePrivate(CountryAttribute)
}

func (b *userBuilderImpl) Email(value string) UserBuilderCanMakeAttributePrivate {
	b.email = ldvalue.NewOptionalString(value)
	return b.canMakeAttributePrivate(EmailAttribute)
}

func (b *userBuilderImpl) FirstName(value string) UserBuilderCanMakeAttributePrivate {
	b.firstName = ldvalue.NewOptionalString(value)
	return b.canMakeAttributePrivate(FirstNameAttribute)
}

func (b *userBuilderImpl) LastName(value string) UserBuilderCanMakeAttributePrivate {
	b.lastName = ldvalue.NewOptionalString(value)
	return b.canMakeAttributePrivate(LastNameAttribute)
}

func (b *userBuilderImpl) Avatar(value string) UserBuilderCanMakeAttributePrivate {
	b.avatar = ldvalue.NewOptionalString(value)
	return b.canMakeAttributePrivate(AvatarAttribute)
}

func (b *userBuilderImpl) Name(value string) UserBuilderCanMakeAttributePrivate {
	b.name = ldvalue.NewOptionalString(value)
	return b.canMakeAttributePrivate(NameAttribute)
}

func (b *userBuilderImpl) Anonymous(value bool) UserBuilder {
	b.anonymous = ldvalue.NewOptionalBool(value)
	return b
}

func (b *userBuilderImpl) Custom(attribute string, value ldvalue.Value) UserBuilderCanMakeAttributePrivate {
	if b.custom == nil {
		b.custom = ldvalue.ValueMapBuild()
	}
	b.custom.Set(attribute, value)
	return b.canMakeAttributePrivate(UserAttribute(attribute))
}

func (b *userBuilderImpl) CustomAll(valueMap ldvalue.ValueMap) UserBuilderCanMakeAttributePrivate {
	if valueMap.Count() == 0 {
		b.custom = nil
	} else {
		b.custom = ldvalue.ValueMapBuildFromMap(valueMap)
	}
	b.lastAttributeCanMakePrivate = ""
	return b
}

func (b *userBuilderImpl) SetAttribute(
	attribute UserAttribute,
	value ldvalue.Value,
) UserBuilderCanMakeAttributePrivate {
	okPrivate := true
	setOptString := func(s *ldvalue.OptionalString) {
		if value.IsString() {
			*s = ldvalue.NewOptionalString(value.StringValue())
		} else if value.IsNull() {
			*s = ldvalue.OptionalString{}
		}
	}
	switch attribute {
	case KeyAttribute:
		if value.IsString() {
			b.key = value.StringValue()
		}
		okPrivate = false
	case SecondaryKeyAttribute:
		setOptString(&b.secondary)
	case IPAttribute:
		setOptString(&b.ip)
	case CountryAttribute:
		setOptString(&b.country)
	case EmailAttribute:
		setOptString(&b.email)
	case FirstNameAttribute:
		setOptString(&b.firstName)
	case LastNameAttribute:
		setOptString(&b.lastName)
	case AvatarAttribute:
		setOptString(&b.avatar)
	case NameAttribute:
		setOptString(&b.name)
	case AnonymousAttribute:
		switch {
		case value.IsNull():
			b.anonymous = ldvalue.OptionalBool{}
		case value.IsBool():
			b.anonymous = ldvalue.NewOptionalBool(value.BoolValue())
		}
		okPrivate = false
	default:
		return b.Custom(string(attribute), value)
	}
	if okPrivate {
		return b.canMakeAttributePrivate(attribute)
	}
	b.lastAttributeCanMakePrivate = ""
	return b
}

func (b *userBuilderImpl) Build() User {
	u := User{
		key:               b.key,
		secondary:         b.secondary,
		ip:                b.ip,
		country:           b.country,
		email:             b.email,
		firstName:         b.firstName,
		lastName:          b.lastName,
		avatar:            b.avatar,
		name:              b.name,
		anonymous:         b.anonymous,
		privateAttributes: b.privateAttrs,
	}
	if b.custom != nil {
		u.custom = b.custom.Build()
	}
	b.privateAttrsCopyOnWrite = true
	return u
}

func (b *userBuilderImpl) AsPrivateAttribute() UserBuilder {
	if b.lastAttributeCanMakePrivate != "" {
		if b.privateAttrs == nil {
			b.privateAttrs = make(map[UserAttribute]struct{})
		} else if b.privateAttrsCopyOnWrite {
			copied := make(map[UserAttribute]struct{}, len(b.privateAttrs))
			for name := range b.privateAttrs {
				copied[name] = struct{}{}
			}
			b.privateAttrs = copied
		}
		b.privateAttrs[b.lastAttributeCanMakePrivate] = struct{}{}
		b.privateAttrsCopyOnWrite = false
	}
	return b
}

func (b *userBuilderImpl) AsNonPrivateAttribute() UserBuilder {
	if b.lastAttributeCanMakePrivate != "" {
		if b.privateAttrs != nil {
			delete(b.privateAttrs, b.lastAttributeCanMakePrivate)
		}
	}
	return b
}
